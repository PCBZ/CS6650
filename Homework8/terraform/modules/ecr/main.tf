# Create (or ensure) an ECR repo exists
resource "aws_ecr_repository" "this" {
  name                 = var.repository_name
  force_delete         = true  # Allow deletion even when images exist
  image_tag_mutability = "MUTABLE"
  
  image_scanning_configuration {
    scan_on_push = false
  }
}
