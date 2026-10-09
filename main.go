package main

import (
	"github.com/muharik28/inventory-reservation-backend/cmd"
	_ "github.com/muharik28/inventory-reservation-backend/cmd/rest"
	_ "github.com/muharik28/inventory-reservation-backend/cmd/rest/scheduler"
)

func main() {
	cmd.Execute()
}
