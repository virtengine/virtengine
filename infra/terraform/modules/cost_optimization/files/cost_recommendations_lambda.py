"""
VirtEngine Cost Recommendations Lambda

This Lambda function generates monthly cost optimization recommendations:
- Right-sizing recommendations from AWS Cost Explorer
- Reserved Instance utilization
- Savings Plans coverage
- Cost trends analysis

Reports are sent via SNS and can be used for monthly cost reviews.
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

# AWS clients
ce = boto3.client('ce')
sns = boto3.client('sns')


def get_date_range(days: int = 30) -> tuple[str, str]:
    """Get date range for Cost Explorer queries."""
    end_date = datetime.now(timezone.utc).date()
    start_date = end_date - timedelta(days=days)
    return start_date.isoformat(), end_date.isoformat()


def get_cost_and_usage(start_date: str, end_date: str) -> dict[str, Any]:
    """Get cost and usage data from Cost Explorer."""
    try:
        response = ce.get_cost_and_usage(
            TimePeriod={
                'Start': start_date,
                'End': end_date
            },
            Granularity='MONTHLY',
            Metrics=['BlendedCost', 'UnblendedCost', 'UsageQuantity'],
            GroupBy=[
                {'Type': 'DIMENSION', 'Key': 'SERVICE'}
            ],
            Filter={
                'Tags': {
                    'Key': 'Project',
                    'Values': ['virtengine']
                }
            }
        )
        return response
    except ClientError as e:
        logger.error(f"Failed to get cost and usage: {e}")
        return {}


def get_rightsizing_recommendations() -> list[dict[str, Any]]:
    """Get EC2 right-sizing recommendations."""
    recommendations = []
    
    try:
        response = ce.get_rightsizing_recommendation(
            Service='AmazonEC2',
            Configuration={
                'RecommendationTarget': 'SAME_INSTANCE_FAMILY',
                'BenefitsConsidered': True
            }
        )
        
        for rec in response.get('RightsizingRecommendations', []):
            current = rec.get('CurrentInstance', {})
            target = rec.get('ModifyRecommendationDetail', {}).get('TargetInstances', [{}])[0]
            
            recommendations.append({
                'account_id': rec.get('AccountId', 'N/A'),
                'instance_id': current.get('ResourceId', 'N/A'),
                'current_type': current.get('ResourceDetails', {}).get('EC2ResourceDetails', {}).get('InstanceType', 'N/A'),
                'recommended_type': target.get('ResourceDetails', {}).get('EC2ResourceDetails', {}).get('InstanceType', 'N/A'),
                'estimated_monthly_savings': target.get('EstimatedMonthlySavings', {}).get('Value', '0'),
                'recommendation_type': rec.get('RightsizingType', 'N/A')
            })
            
    except ClientError as e:
        logger.error(f"Failed to get rightsizing recommendations: {e}")
    
    return recommendations


def get_reservation_utilization() -> dict[str, Any]:
    """Get Reserved Instance utilization."""
    start_date, end_date = get_date_range(30)
    
    try:
        response = ce.get_reservation_utilization(
            TimePeriod={
                'Start': start_date,
                'End': end_date
            },
            Granularity='MONTHLY'
        )
        
        utilization = response.get('UtilizationsByTime', [{}])[0]
        total = utilization.get('Total', {})
        
        return {
            'utilization_percentage': total.get('UtilizationPercentage', '0'),
            'purchased_hours': total.get('PurchasedHours', '0'),
            'used_hours': total.get('TotalActualHours', '0'),
            'unused_hours': total.get('UnusedHours', '0'),
            'on_demand_cost_equivalent': total.get('OnDemandCostOfRIHoursUsed', '0'),
            'net_savings': total.get('NetRISavings', '0')
        }
    except ClientError as e:
        logger.error(f"Failed to get reservation utilization: {e}")
        return {}


def get_savings_plans_coverage() -> dict[str, Any]:
    """Get Savings Plans coverage."""
    start_date, end_date = get_date_range(30)
    
    try:
        response = ce.get_savings_plans_coverage(
            TimePeriod={
                'Start': start_date,
                'End': end_date
            },
            Granularity='MONTHLY'
        )
        
        coverage = response.get('SavingsPlansCoverages', [{}])[0]
        
        return {
            'coverage_percentage': coverage.get('Coverage', {}).get('CoveragePercentage', '0'),
            'spend_covered': coverage.get('Coverage', {}).get('SpendCoveredBySavingsPlans', '0'),
            'on_demand_cost': coverage.get('Coverage', {}).get('OnDemandCost', '0')
        }
    except ClientError as e:
        logger.error(f"Failed to get savings plans coverage: {e}")
        return {}


def get_savings_plans_utilization() -> dict[str, Any]:
    """Get Savings Plans utilization."""
    start_date, end_date = get_date_range(30)
    
    try:
        response = ce.get_savings_plans_utilization(
            TimePeriod={
                'Start': start_date,
                'End': end_date
            },
            Granularity='MONTHLY'
        )
        
        utilization = response.get('SavingsPlansUtilizationsByTime', [{}])[0]
        
        return {
            'utilization_percentage': utilization.get('Utilization', {}).get('UtilizationPercentage', '0'),
            'used_commitment': utilization.get('Utilization', {}).get('UsedCommitment', '0'),
            'unused_commitment': utilization.get('Utilization', {}).get('UnusedCommitment', '0'),
            'total_commitment': utilization.get('Utilization', {}).get('TotalCommitment', '0')
        }
    except ClientError as e:
        logger.error(f"Failed to get savings plans utilization: {e}")
        return {}


def analyze_cost_trends(cost_data: dict[str, Any]) -> dict[str, Any]:
    """Analyze cost trends from cost and usage data."""
    results_by_time = cost_data.get('ResultsByTime', [])
    
    if len(results_by_time) < 2:
        return {'trend': 'insufficient_data'}
    
    # Calculate month-over-month change
    services = {}
    for period in results_by_time:
        for group in period.get('Groups', []):
            service = group.get('Keys', ['Unknown'])[0]
            cost = float(group.get('Metrics', {}).get('BlendedCost', {}).get('Amount', 0))
            
            if service not in services:
                services[service] = []
            services[service].append(cost)
    
    trends = {}
    for service, costs in services.items():
        if len(costs) >= 2:
            change = costs[-1] - costs[-2]
            pct_change = (change / costs[-2] * 100) if costs[-2] > 0 else 0
            trends[service] = {
                'previous_month': round(costs[-2], 2),
                'current_month': round(costs[-1], 2),
                'change': round(change, 2),
                'pct_change': round(pct_change, 1)
            }
    
    # Sort by absolute change
    sorted_trends = dict(sorted(trends.items(), key=lambda x: abs(x[1]['change']), reverse=True))
    
    return sorted_trends


def generate_recommendations(
    rightsizing: list,
    ri_utilization: dict,
    sp_coverage: dict,
    sp_utilization: dict,
    trends: dict
) -> list[str]:
    """Generate actionable recommendations based on analysis."""
    recommendations = []
    
    # Right-sizing recommendations
    total_rightsizing_savings = sum(float(r.get('estimated_monthly_savings', 0)) for r in rightsizing)
    if total_rightsizing_savings > 0:
        recommendations.append(
            f"Right-size {len(rightsizing)} EC2 instances for estimated savings of ${total_rightsizing_savings:.2f}/month"
        )
    
    # RI utilization recommendations
    ri_util = float(ri_utilization.get('utilization_percentage', '100'))
    if ri_util < 80:
        unused = float(ri_utilization.get('unused_hours', '0'))
        recommendations.append(
            f"Reserved Instance utilization is {ri_util:.1f}%. Consider selling unused RIs or right-sizing workloads. "
            f"Unused hours: {unused}"
        )
    
    # Savings Plans coverage
    sp_cov = float(sp_coverage.get('coverage_percentage', '0'))
    if sp_cov < 70:
        on_demand = float(sp_coverage.get('on_demand_cost', '0'))
        recommendations.append(
            f"Savings Plans coverage is {sp_cov:.1f}%. On-demand spend: ${on_demand:.2f}. "
            "Consider purchasing Savings Plans for predictable workloads."
        )
    
    # Savings Plans utilization
    sp_util = float(sp_utilization.get('utilization_percentage', '100'))
    if sp_util < 90:
        unused = float(sp_utilization.get('unused_commitment', '0'))
        recommendations.append(
            f"Savings Plans utilization is {sp_util:.1f}%. Unused commitment: ${unused:.2f}"
        )
    
    # Cost trend recommendations
    for service, trend in list(trends.items())[:5]:  # Top 5 by change
        if trend['pct_change'] > 20:
            recommendations.append(
                f"{service}: Cost increased by {trend['pct_change']:.1f}% (${trend['change']:.2f}). "
                "Review usage patterns."
            )
    
    if not recommendations:
        recommendations.append("Infrastructure costs are well-optimized. Continue monitoring for opportunities.")
    
    return recommendations


def format_report(
    cost_data: dict,
    rightsizing: list,
    ri_utilization: dict,
    sp_coverage: dict,
    sp_utilization: dict,
    trends: dict,
    recommendations: list
) -> str:
    """Format the cost recommendations report."""
    report = f"""
================================================================================
VirtEngine Monthly Cost Optimization Report
================================================================================
Environment: {ENVIRONMENT}
Report Date: {datetime.now(timezone.utc).strftime('%Y-%m-%d %H:%M:%S UTC')}

--------------------------------------------------------------------------------
EXECUTIVE SUMMARY
--------------------------------------------------------------------------------

"""

    # Add top services by cost
    report += "Top Services by Cost (Last 30 Days):\n"
    if trends:
        for i, (service, data) in enumerate(list(trends.items())[:10], 1):
            report += f"  {i}. {service}: ${data['current_month']:.2f}"
            if data['pct_change'] != 0:
                direction = "↑" if data['pct_change'] > 0 else "↓"
                report += f" ({direction}{abs(data['pct_change']):.1f}%)"
            report += "\n"
    
    # Reserved Instances
    report += f"""
--------------------------------------------------------------------------------
RESERVED INSTANCES
--------------------------------------------------------------------------------
Utilization: {ri_utilization.get('utilization_percentage', 'N/A')}%
Purchased Hours: {ri_utilization.get('purchased_hours', 'N/A')}
Used Hours: {ri_utilization.get('used_hours', 'N/A')}
Unused Hours: {ri_utilization.get('unused_hours', 'N/A')}
Net Savings: ${float(ri_utilization.get('net_savings', 0)):.2f}
"""

    # Savings Plans
    report += f"""
--------------------------------------------------------------------------------
SAVINGS PLANS
--------------------------------------------------------------------------------
Coverage: {sp_coverage.get('coverage_percentage', 'N/A')}%
Utilization: {sp_utilization.get('utilization_percentage', 'N/A')}%
Used Commitment: ${float(sp_utilization.get('used_commitment', 0)):.2f}
Unused Commitment: ${float(sp_utilization.get('unused_commitment', 0)):.2f}
"""

    # Right-sizing
    if rightsizing:
        total_savings = sum(float(r.get('estimated_monthly_savings', 0)) for r in rightsizing)
        report += f"""
--------------------------------------------------------------------------------
RIGHT-SIZING RECOMMENDATIONS
--------------------------------------------------------------------------------
Total Recommendations: {len(rightsizing)}
Estimated Monthly Savings: ${total_savings:.2f}

Top Recommendations:
"""
        for rec in rightsizing[:5]:
            report += f"  - {rec['instance_id']}: {rec['current_type']} → {rec['recommended_type']} "
            report += f"(saves ${float(rec['estimated_monthly_savings']):.2f}/mo)\n"

    # Actionable recommendations
    report += """
--------------------------------------------------------------------------------
ACTIONABLE RECOMMENDATIONS
--------------------------------------------------------------------------------
"""
    for i, rec in enumerate(recommendations, 1):
        report += f"{i}. {rec}\n"

    report += """
--------------------------------------------------------------------------------
NEXT STEPS
--------------------------------------------------------------------------------
1. Review right-sizing recommendations and implement changes
2. Evaluate Reserved Instance and Savings Plans purchases
3. Investigate services with cost increases > 20%
4. Schedule monthly cost review meeting

For questions, contact the platform team.
================================================================================
"""
    
    return report


def send_report(report: str) -> None:
    """Send the report via SNS."""
    if not SNS_TOPIC_ARN:
        logger.warning("SNS_TOPIC_ARN not configured")
        return
    
    try:
        sns.publish(
            TopicArn=SNS_TOPIC_ARN,
            Subject=f"VirtEngine Monthly Cost Report [{ENVIRONMENT}]",
            Message=report
        )
        logger.info("Cost report sent via SNS")
    except ClientError as e:
        logger.error(f"Failed to send report: {e}")


def handler(event: dict[str, Any], context: Any) -> dict[str, Any]:
    """Main Lambda handler."""
    logger.info(f"Generating cost recommendations for: {ENVIRONMENT}")
    
    # Gather data
    start_date, end_date = get_date_range(60)  # Last 60 days for trends
    
    cost_data = get_cost_and_usage(start_date, end_date)
    rightsizing = get_rightsizing_recommendations()
    ri_utilization = get_reservation_utilization()
    sp_coverage = get_savings_plans_coverage()
    sp_utilization = get_savings_plans_utilization()
    trends = analyze_cost_trends(cost_data)
    
    # Generate recommendations
    recommendations = generate_recommendations(
        rightsizing, ri_utilization, sp_coverage, sp_utilization, trends
    )
    
    # Format and send report
    report = format_report(
        cost_data, rightsizing, ri_utilization, sp_coverage,
        sp_utilization, trends, recommendations
    )
    
    logger.info(f"Generated report:\n{report}")
    send_report(report)
    
    return {
        'statusCode': 200,
        'body': json.dumps({
            'message': 'Cost recommendations generated successfully',
            'recommendations_count': len(recommendations),
            'rightsizing_count': len(rightsizing)
        })
    }
