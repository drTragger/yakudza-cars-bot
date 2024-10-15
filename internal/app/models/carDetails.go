package models

type PriceRange struct {
	Title string
	Min   int
	Max   int
}

type CarDetails struct {
	Phone string
	Price *PriceRange
	Year  string
}
