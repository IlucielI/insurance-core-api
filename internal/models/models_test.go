package models

import "testing"

func TestTableNames(t *testing.T) {
	product := Product{}
	if product.TableName() != "products" {
		t.Fatalf("Product.TableName() = %q, want products", product.TableName())
	}

	application := Application{}
	if application.TableName() != "applications" {
		t.Fatalf("Application.TableName() = %q, want applications", application.TableName())
	}
}
