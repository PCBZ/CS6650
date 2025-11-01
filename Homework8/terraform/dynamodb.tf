
resource "aws_dynamodb_table" "shopping_carts" {
  name         = "shopping_carts"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "cart_id"
  attribute {
    name = "cart_id"
    type = "N"
  }
  attribute {
    name = "customer_id"
    type = "N"
  }
  global_secondary_index {
    name            = "customer_id-index"
    hash_key        = "customer_id"
    projection_type = "ALL"
  }
}

resource "aws_dynamodb_table" "shopping_cart_items" {
  name         = "shopping_cart_items"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "cart_id"
  range_key    = "item_id"
  attribute {
    name = "cart_id"
    type = "N"
  }
  attribute {
    name = "item_id"
    type = "S"
  }
}
