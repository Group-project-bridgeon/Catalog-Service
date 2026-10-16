package router

// PermissionMatrix enumerates every supported business route. Status changes
// use update; nested collections use the permission for the returned resource.
func PermissionMatrix() map[string][]string {
	policy := make(map[string][]string)
	const base = "/api/v1/catalog/"
	for _, resource := range []string{"categories", "brands", "units", "products", "product-variants", "attributes", "attribute-values", "variant-values", "product-images", "price-lists", "prices"} {
		path := base + resource
		item := "/:id"
		if resource == "variant-values" {
			item = "/:variant_id/:attribute_id"
		}
		policy["GET "+path] = []string{"catalog." + resource + ":read"}
		policy["GET "+path+item] = []string{"catalog." + resource + ":read"}
		policy["POST "+path] = []string{"catalog." + resource + ":create"}
		policy["PUT "+path+item] = []string{"catalog." + resource + ":update"}
	}
	for _, resource := range []string{"categories", "brands", "units", "products", "product-variants", "price-lists"} {
		policy["PATCH "+base+resource+"/:id/status"] = []string{"catalog." + resource + ":update"}
	}
	for path, resource := range map[string]string{
		"attributes/:id/values":       "attribute-values",
		"products/:id/variants":       "product-variants",
		"product-variants/:id/values": "variant-values",
		"products/:id/images":         "product-images",
		"product-variants/:id/images": "product-images",
		"price-lists/:id/prices":      "prices",
		"product-variants/:id/prices": "prices",
	} {
		policy["GET "+base+path] = []string{"catalog." + resource + ":read"}
	}
	policy["GET "+base+"products/:id/detail"] = []string{
		"catalog.products:read", "catalog.product-variants:read", "catalog.variant-values:read", "catalog.product-images:read", "catalog.prices:read",
	}
	return policy
}
