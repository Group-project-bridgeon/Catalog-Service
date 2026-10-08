package domain

type ProductDetail struct {
	Product       Product
	Variants      []ProductVariant
	VariantValues []VariantValue
	Images        []ProductImage
	Prices        []Price
}
