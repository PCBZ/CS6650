
resource "aws_dynamodb_table" "shopping_carts" {
  name         = "shopping_carts"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "cart_id"
  attribute {
    name = "cart_id"
    type = "N"
  }
}

resource "aws_dynamodb_table" "shopping_cart_items" {
  name         = "shopping_cart_items"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "cart_id"
  range_key    = "product_id"
  attribute {
    name = "cart_id"
    type = "N"
  }
  attribute {
    name = "product_id"
    type = "N"
  }
}
