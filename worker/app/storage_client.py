import logging
import os

import boto3
from botocore.client import Config
from botocore.exceptions import ClientError

log = logging.getLogger(__name__)

S3_ENDPOINT = os.environ.get("S3_ENDPOINT", "http://localhost:9000")
S3_ACCESS_KEY = os.environ.get("S3_ACCESS_KEY", "minio")
S3_SECRET_KEY = os.environ.get("S3_SECRET_KEY", "password")
S3_BUCKET = os.environ.get("S3_BUCKET", "images")

class Storage:
    """A thin wrapper around the S3 API for one bucket."""

    def __init__(
        self,
        endpoint=S3_ENDPOINT,
        access_key=S3_ACCESS_KEY,
        secret_key=S3_SECRET_KEY,
        bucket=S3_BUCKET,
    ):
        self.bucket = bucket
        self.client = boto3.client(
            "s3", #
            endpoint_url=endpoint,
            aws_access_key_id=access_key,
            aws_secret_access_key=secret_key,
            # MinIO serves buckets as paths (host/bucket/key), not as
            # subdomains (bucket.host/key) like AWS does.
            config=Config(signature_version="s3v4", s3={"addressing_style": "path"}),
        )

    def download(self, key):
        """Fetch an object and return its raw bytes."""
        response = self.client.get_object(Bucket=self.bucket, Key=key)
        return response["Body"].read()

    def upload(self, key, data, content_type="image/png"):
        """Store raw bytes under the given key."""
        self.client.put_object(
            Bucket=self.bucket, Key=key, Body=data, ContentType=content_type
        )
        log.info("uploaded %s (%d bytes)", key, len(data))

    def ensure_bucket(self):
        """Create the bucket if it doesn't exist yet (dev convenience)."""
        try:
            self.client.head_bucket(Bucket=self.bucket)
        except ClientError:
            log.info("creating bucket %r", self.bucket)
            self.client.create_bucket(Bucket=self.bucket)


if __name__ == "__main__":
    # Standalone smoke test against the dev MinIO: round-trip an object.
    logging.basicConfig(level=logging.INFO)
    storage = Storage()
    storage.ensure_bucket()
    storage.upload("smoke-test.txt", b"hello", content_type="text/plain")
    assert storage.download("smoke-test.txt") == b"hello"
    print("storage round-trip OK")
