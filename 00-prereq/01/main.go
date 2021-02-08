// hands-on 01
package main

import (
	"fmt"
	"math"
)

// create a type square
type square struct {
	side float64
}

// create a type circle
type circle struct {
	radius float64
}

// attach a method to each that calculates area and returns it
func (s square) area() float64 {
	return math.Pow(s.side, 2)
}

func (c circle) area() float64 {
	return math.Pi * math.Pow(c.radius, 2)
}

// create a type shape which defines an interface as anything which has the area method
type shape interface {
	area() float64
}

// create a func info which takes type shape and then prints the area
func shapeArea(s shape) float64 {
	return s.area()
}

func main() {
	s := square{5}
	c := circle{5}
	fmt.Println(s.area())
	fmt.Println(c.area())
	fmt.Println(shapeArea(s))
	fmt.Println(shapeArea(c))
}
