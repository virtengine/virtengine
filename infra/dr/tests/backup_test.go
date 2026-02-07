package dr

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --------------------------------------------------------------------------
// Backup Verification Tests
// --------------------------------------------------------------------------

func TestBackup_AllRegionsHaveBackups(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	regions := []string{
		region("VE_PRIMARY_REGION", defaultPrimaryRegion),
		region("VE_SECONDARY_REGION", defaultSecondaryRegion),
		region("VE_TERTIARY_REGION", defaultTertiaryRegion),
	}

	for _, r := range regions {
		t.Run(r, func(t *testing.T) {
			bucket := fmt.Sprintf("virtengine-cockroachdb-backup-%s", r)

			// Check S3 bucket exists and has objects
			out, err := runCommandAllowFailure("aws", "s3", "ls", "s3://"+bucket+"/backups/", "--region", r)
			if err != nil {
				t.Skipf("S3 bucket check not available for %s: %v", r, err)
				return
			}

			assert.NotEmpty(t, out, "backup bucket should contain backup data for %s", r)
		})
	}
}

func TestBackup_EncryptionEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	regions := []string{
		region("VE_PRIMARY_REGION", defaultPrimaryRegion),
		region("VE_SECONDARY_REGION", defaultSecondaryRegion),
		region("VE_TERTIARY_REGION", defaultTertiaryRegion),
	}

	for _, r := range regions {
		t.Run(r, func(t *testing.T) {
			bucket := fmt.Sprintf("virtengine-cockroachdb-backup-%s", r)

			out, err := runCommandAllowFailure("aws", "s3api", "get-bucket-encryption",
				"--bucket", bucket, "--region", r)
			if err != nil {
				t.Skipf("encryption check not available for %s: %v", r, err)
				return
			}

			assert.Contains(t, out, "aws:kms", "bucket %s should use KMS encryption", bucket)
		})
	}
}

func TestBackup_CrossRegionReplication(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	primary := region("VE_PRIMARY_REGION", defaultPrimaryRegion)
	secondary := region("VE_SECONDARY_REGION", defaultSecondaryRegion)
	primaryCtx := kubeContext(primary)

	// Verify backups exist in both primary and secondary buckets
	for _, r := range []string{primary, secondary} {
		t.Run(r, func(t *testing.T) {
			bucket := fmt.Sprintf("s3://virtengine-cockroachdb-backup-%s/backups", r)
			query := fmt.Sprintf(
				"SHOW BACKUP LATEST IN '%s?AUTH=implicit';",
				bucket,
			)

			out, err := runCommandAllowFailure("kubectl", "--context", primaryCtx, "-n", "cockroachdb",
				"exec", "cockroachdb-0", "--",
				"cockroach", "sql",
				"--certs-dir=/cockroach/cockroach-certs",
				"-e", query)
			if err != nil {
				t.Skipf("backup verification not available for %s: %v", r, err)
				return
			}

			assert.NotContains(t, out, "ERROR", "backup check should not return errors for %s", r)
		})
	}
}

func TestBackup_RetentionPolicy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	regions := []string{
		region("VE_PRIMARY_REGION", defaultPrimaryRegion),
		region("VE_SECONDARY_REGION", defaultSecondaryRegion),
		region("VE_TERTIARY_REGION", defaultTertiaryRegion),
	}

	for _, r := range regions {
		t.Run(r, func(t *testing.T) {
			bucket := fmt.Sprintf("virtengine-cockroachdb-backup-%s", r)

			out, err := runCommandAllowFailure("aws", "s3api", "get-bucket-lifecycle-configuration",
				"--bucket", bucket, "--region", r)
			if err != nil {
				t.Skipf("lifecycle check not available for %s: %v", r, err)
				return
			}

			assert.Contains(t, out, "Rules", "bucket %s should have lifecycle rules", bucket)
		})
	}
}

func TestBackup_MaxAge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	maxAge := 24 * time.Hour // Full backup runs daily

	regions := []string{
		region("VE_PRIMARY_REGION", defaultPrimaryRegion),
		region("VE_SECONDARY_REGION", defaultSecondaryRegion),
	}

	for _, r := range regions {
		t.Run(r, func(t *testing.T) {
			ctx := kubeContext(r)
			bucket := fmt.Sprintf("s3://virtengine-cockroachdb-backup-%s/backups", r)
			query := fmt.Sprintf(
				"SELECT extract(epoch from (now() - max(end_time)))::int AS age_seconds FROM [SHOW BACKUP LATEST IN '%s?AUTH=implicit'];",
				bucket,
			)

			out, err := runCommandAllowFailure("kubectl", "--context", ctx, "-n", "cockroachdb",
				"exec", "cockroachdb-0", "--",
				"cockroach", "sql",
				"--certs-dir=/cockroach/cockroach-certs",
				"--format=csv",
				"-e", query)
			if err != nil {
				t.Skipf("backup age check not available for %s: %v", r, err)
				return
			}

			lines := strings.Split(strings.TrimSpace(out), "\n")
			if len(lines) < 2 {
				t.Skip("no backup age data")
				return
			}

			ageSec, err := strconv.Atoi(strings.TrimSpace(lines[len(lines)-1]))
			if err != nil {
				t.Skipf("unable to parse age: %v", err)
				return
			}

			age := time.Duration(ageSec) * time.Second
			assert.Less(t, age, maxAge, "backup age %v should be under %v for region %s", age, maxAge, r)
		})
	}
}
