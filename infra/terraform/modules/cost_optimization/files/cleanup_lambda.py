"""
VirtEngine Unused Resource Cleanup Lambda

This Lambda function identifies and cleans up unused AWS resources:
- Unattached EBS volumes
- Old EBS snapshots
- Unused Elastic IPs
- Old AMIs

It runs on a schedule and sends cleanup reports via SNS.
"""

import os
import json
import logging
from datetime import datetime, timedelta, timezone
from typing import Any

import boto3
from botocore.exceptions import ClientError

# Configure logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Environment variables
ENVIRONMENT = os.environ.get('ENVIRONMENT', 'dev')
SNS_TOPIC_ARN = os.environ.get('SNS_TOPIC_ARN', '')
DRY_RUN = os.environ.get('DRY_RUN', 'true').lower() == 'true'
SNAPSHOT_RETENTION_DAYS = int(os.environ.get('SNAPSHOT_RETENTION_DAYS', '30'))
VOLUME_RETENTION_DAYS = int(os.environ.get('VOLUME_RETENTION_DAYS', '7'))
AMI_RETENTION_DAYS = int(os.environ.get('AMI_RETENTION_DAYS', '90'))
EIP_RETENTION_DAYS = int(os.environ.get('EIP_RETENTION_DAYS', '3'))

# AWS clients
ec2 = boto3.client('ec2')
sns = boto3.client('sns')


def get_unattached_volumes() -> list[dict[str, Any]]:
    """Find EBS volumes that are not attached to any instance."""
    volumes = []
    paginator = ec2.get_paginator('describe_volumes')
    
    for page in paginator.paginate(
        Filters=[
            {'Name': 'status', 'Values': ['available']},
            {'Name': 'tag:Project', 'Values': ['virtengine']},
            {'Name': 'tag:Environment', 'Values': [ENVIRONMENT]}
        ]
    ):
        for volume in page['Volumes']:
            create_time = volume['CreateTime']
            age_days = (datetime.now(timezone.utc) - create_time).days
            
            if age_days >= VOLUME_RETENTION_DAYS:
                volumes.append({
                    'VolumeId': volume['VolumeId'],
                    'Size': volume['Size'],
                    'CreateTime': create_time.isoformat(),
                    'AgeDays': age_days,
                    'Tags': volume.get('Tags', [])
                })
    
    return volumes


def get_old_snapshots() -> list[dict[str, Any]]:
    """Find EBS snapshots older than retention period."""
    snapshots = []
    account_id = boto3.client('sts').get_caller_identity()['Account']
    paginator = ec2.get_paginator('describe_snapshots')
    
    for page in paginator.paginate(
        OwnerIds=[account_id],
        Filters=[
            {'Name': 'tag:Project', 'Values': ['virtengine']},
            {'Name': 'tag:Environment', 'Values': [ENVIRONMENT]}
        ]
    ):
        for snapshot in page['Snapshots']:
            start_time = snapshot['StartTime']
            age_days = (datetime.now(timezone.utc) - start_time).days
            
            if age_days >= SNAPSHOT_RETENTION_DAYS:
                # Check if snapshot is used by any AMI
                if not is_snapshot_in_use(snapshot['SnapshotId']):
                    snapshots.append({
                        'SnapshotId': snapshot['SnapshotId'],
                        'VolumeId': snapshot.get('VolumeId', 'N/A'),
                        'VolumeSize': snapshot['VolumeSize'],
                        'StartTime': start_time.isoformat(),
                        'AgeDays': age_days,
                        'Description': snapshot.get('Description', ''),
                        'Tags': snapshot.get('Tags', [])
                    })
    
    return snapshots


def is_snapshot_in_use(snapshot_id: str) -> bool:
    """Check if a snapshot is used by any AMI."""
    try:
        response = ec2.describe_images(
            Filters=[
                {'Name': 'block-device-mapping.snapshot-id', 'Values': [snapshot_id]}
            ]
        )
        return len(response['Images']) > 0
    except ClientError:
        return True  # Assume in use if we can't check


def get_unused_eips() -> list[dict[str, Any]]:
    """Find Elastic IPs that are not associated with any resource."""
    eips = []
    
    response = ec2.describe_addresses(
        Filters=[
            {'Name': 'tag:Project', 'Values': ['virtengine']},
            {'Name': 'tag:Environment', 'Values': [ENVIRONMENT]}
        ]
    )
    
    for eip in response['Addresses']:
        # EIP is unused if it has no association
        if 'AssociationId' not in eip and 'InstanceId' not in eip:
            # Check allocation time from tags or use a default
            tags = {t['Key']: t['Value'] for t in eip.get('Tags', [])}
            allocation_date = tags.get('AllocationDate', '')
            
            eips.append({
                'AllocationId': eip.get('AllocationId', ''),
                'PublicIp': eip['PublicIp'],
                'Domain': eip['Domain'],
                'Tags': eip.get('Tags', [])
            })
    
    return eips


def get_old_amis() -> list[dict[str, Any]]:
    """Find AMIs older than retention period."""
    amis = []
    account_id = boto3.client('sts').get_caller_identity()['Account']
    
    response = ec2.describe_images(
        Owners=[account_id],
        Filters=[
            {'Name': 'tag:Project', 'Values': ['virtengine']},
            {'Name': 'tag:Environment', 'Values': [ENVIRONMENT]}
        ]
    )
    
    for image in response['Images']:
        creation_date = datetime.fromisoformat(image['CreationDate'].replace('Z', '+00:00'))
        age_days = (datetime.now(timezone.utc) - creation_date).days
        
        if age_days >= AMI_RETENTION_DAYS:
            # Check if AMI is in use by any running instance
            if not is_ami_in_use(image['ImageId']):
                amis.append({
                    'ImageId': image['ImageId'],
                    'Name': image.get('Name', 'N/A'),
                    'CreationDate': image['CreationDate'],
                    'AgeDays': age_days,
                    'Tags': image.get('Tags', [])
                })
    
    return amis


def is_ami_in_use(ami_id: str) -> bool:
    """Check if an AMI is used by any running instance."""
    try:
        response = ec2.describe_instances(
            Filters=[
                {'Name': 'image-id', 'Values': [ami_id]},
                {'Name': 'instance-state-name', 'Values': ['pending', 'running', 'stopping', 'stopped']}
            ]
        )
        for reservation in response['Reservations']:
            if len(reservation['Instances']) > 0:
                return True
        return False
    except ClientError:
        return True  # Assume in use if we can't check


def delete_volume(volume_id: str) -> bool:
    """Delete an EBS volume."""
    if DRY_RUN:
        logger.info(f"[DRY RUN] Would delete volume: {volume_id}")
        return True
    
    try:
        ec2.delete_volume(VolumeId=volume_id)
        logger.info(f"Deleted volume: {volume_id}")
        return True
    except ClientError as e:
        logger.error(f"Failed to delete volume {volume_id}: {e}")
        return False


def delete_snapshot(snapshot_id: str) -> bool:
    """Delete an EBS snapshot."""
    if DRY_RUN:
        logger.info(f"[DRY RUN] Would delete snapshot: {snapshot_id}")
        return True
    
    try:
        ec2.delete_snapshot(SnapshotId=snapshot_id)
        logger.info(f"Deleted snapshot: {snapshot_id}")
        return True
    except ClientError as e:
        logger.error(f"Failed to delete snapshot {snapshot_id}: {e}")
        return False


def release_eip(allocation_id: str) -> bool:
    """Release an Elastic IP."""
    if DRY_RUN:
        logger.info(f"[DRY RUN] Would release EIP: {allocation_id}")
        return True
    
    try:
        ec2.release_address(AllocationId=allocation_id)
        logger.info(f"Released EIP: {allocation_id}")
        return True
    except ClientError as e:
        logger.error(f"Failed to release EIP {allocation_id}: {e}")
        return False


def deregister_ami(ami_id: str) -> bool:
    """Deregister an AMI."""
    if DRY_RUN:
        logger.info(f"[DRY RUN] Would deregister AMI: {ami_id}")
        return True
    
    try:
        ec2.deregister_image(ImageId=ami_id)
        logger.info(f"Deregistered AMI: {ami_id}")
        return True
    except ClientError as e:
        logger.error(f"Failed to deregister AMI {ami_id}: {e}")
        return False


def calculate_savings(volumes: list, snapshots: list, eips: list, amis: list) -> dict[str, float]:
    """Calculate estimated monthly savings from cleanup."""
    # Approximate costs (USD per month)
    EBS_COST_PER_GB = 0.10  # gp3 average
    SNAPSHOT_COST_PER_GB = 0.05
    EIP_COST_PER_MONTH = 3.60  # Unused EIP cost
    
    volume_savings = sum(v['Size'] for v in volumes) * EBS_COST_PER_GB
    snapshot_savings = sum(s['VolumeSize'] for s in snapshots) * SNAPSHOT_COST_PER_GB
    eip_savings = len(eips) * EIP_COST_PER_MONTH
    
    return {
        'volumes': round(volume_savings, 2),
        'snapshots': round(snapshot_savings, 2),
        'eips': round(eip_savings, 2),
        'total': round(volume_savings + snapshot_savings + eip_savings, 2)
    }


def send_report(report: dict[str, Any]) -> None:
    """Send cleanup report via SNS."""
    if not SNS_TOPIC_ARN:
        logger.warning("SNS_TOPIC_ARN not configured, skipping notification")
        return
    
    message = f"""
VirtEngine Resource Cleanup Report
===================================
Environment: {ENVIRONMENT}
Execution Time: {report['execution_time']}
Mode: {'DRY RUN' if DRY_RUN else 'LIVE'}

Summary:
--------
- Unattached Volumes: {report['summary']['volumes_found']} found, {report['summary']['volumes_deleted']} deleted
- Old Snapshots: {report['summary']['snapshots_found']} found, {report['summary']['snapshots_deleted']} deleted
- Unused EIPs: {report['summary']['eips_found']} found, {report['summary']['eips_released']} released
- Old AMIs: {report['summary']['amis_found']} found, {report['summary']['amis_deregistered']} deregistered

Estimated Monthly Savings:
-------------------------
- Volumes: ${report['savings']['volumes']}
- Snapshots: ${report['savings']['snapshots']}
- EIPs: ${report['savings']['eips']}
- Total: ${report['savings']['total']}

{'Note: This was a DRY RUN. No resources were actually deleted.' if DRY_RUN else 'Resources have been cleaned up.'}

For detailed resource list, check CloudWatch Logs.
"""
    
    try:
        sns.publish(
            TopicArn=SNS_TOPIC_ARN,
            Subject=f"VirtEngine Resource Cleanup Report [{ENVIRONMENT}]",
            Message=message
        )
        logger.info("Cleanup report sent via SNS")
    except ClientError as e:
        logger.error(f"Failed to send SNS notification: {e}")


def handler(event: dict[str, Any], context: Any) -> dict[str, Any]:
    """Main Lambda handler."""
    logger.info(f"Starting resource cleanup for environment: {ENVIRONMENT}")
    logger.info(f"Dry run mode: {DRY_RUN}")
    
    # Collect unused resources
    volumes = get_unattached_volumes()
    snapshots = get_old_snapshots()
    eips = get_unused_eips()
    amis = get_old_amis()
    
    logger.info(f"Found: {len(volumes)} volumes, {len(snapshots)} snapshots, {len(eips)} EIPs, {len(amis)} AMIs")
    
    # Calculate savings
    savings = calculate_savings(volumes, snapshots, eips, amis)
    
    # Perform cleanup
    volumes_deleted = sum(1 for v in volumes if delete_volume(v['VolumeId']))
    snapshots_deleted = sum(1 for s in snapshots if delete_snapshot(s['SnapshotId']))
    eips_released = sum(1 for e in eips if e.get('AllocationId') and release_eip(e['AllocationId']))
    amis_deregistered = sum(1 for a in amis if deregister_ami(a['ImageId']))
    
    # Generate report
    report = {
        'execution_time': datetime.now(timezone.utc).isoformat(),
        'environment': ENVIRONMENT,
        'dry_run': DRY_RUN,
        'summary': {
            'volumes_found': len(volumes),
            'volumes_deleted': volumes_deleted,
            'snapshots_found': len(snapshots),
            'snapshots_deleted': snapshots_deleted,
            'eips_found': len(eips),
            'eips_released': eips_released,
            'amis_found': len(amis),
            'amis_deregistered': amis_deregistered
        },
        'savings': savings,
        'details': {
            'volumes': volumes,
            'snapshots': snapshots,
            'eips': eips,
            'amis': amis
        }
    }
    
    # Log detailed findings
    logger.info(f"Cleanup report: {json.dumps(report, default=str)}")
    
    # Send notification
    send_report(report)
    
    return {
        'statusCode': 200,
        'body': json.dumps({
            'message': 'Cleanup completed successfully',
            'summary': report['summary'],
            'savings': report['savings']
        })
    }
