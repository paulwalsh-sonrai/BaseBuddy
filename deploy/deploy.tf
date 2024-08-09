provider "aws" {
  region = var.aws_region
}

# Variables for networking
# Note: we 
variable "vpc_id" {}
variable "subnet_id" {}
variable "security_group_id" {}

# EC2 instance key pair name
variable "key_name" {}

# Variables for instance
variable "instance_type" {
  default = "t2.micro"
}

# Variables for S3 bucket
variable "bucket_name" {}

# IAM Role for EC2 to access S3
resource "aws_iam_role" "ec2_role" {
  name = "ec2_s3_access_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

# IAM Policy to allow S3 CRUD operations
resource "aws_iam_policy" "s3_crud_policy" {
  name        = "S3CrudPolicy"
  description = "Policy to allow EC2 instance to perform CRUD operations on the S3 bucket"
  policy      = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Effect   = "Allow"
        Resource = [
          "arn:aws:s3:::${var.bucket_name}/*",
          "arn:aws:s3:::${var.bucket_name}"
        ]
      },
    ]
  })
}

# Attach the policy to the role
resource "aws_iam_role_policy_attachment" "attach_s3_crud_policy" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = aws_iam_policy.s3_crud_policy.arn
}

# Instance profile to allow EC2 to use the role
resource "aws_iam_instance_profile" "ec2_instance_profile" {
  name = "ec2_instance_profile"
  role = aws_iam_role.ec2_role.name
}

# EC2 instance
resource "aws_instance" "ec2_instance" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = var.subnet_id
  security_groups        = [var.security_group_id]
  key_name               = var.key_name
  iam_instance_profile   = aws_iam_instance_profile.ec2_instance_profile.name

  # Root volume with 4GB size
  root_block_device {
    volume_size = 4
  }

  tags = {
    Name = "EC2WithS3Access"
  }
}

# Data source to fetch the latest Ubuntu AMI
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}

# S3 Bucket
resource "aws_s3_bucket" "bucket" {
  bucket = var.bucket_name

  tags = {
    Name = "EC2AccessibleBucket"
  }
}

output "ec2_instance_id" {
  value = aws_instance.ec2_instance.id
}

output "s3_bucket_name" {
  value = aws_s3_bucket.bucket.bucket
}
