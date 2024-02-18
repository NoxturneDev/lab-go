package main

import "fmt"

type Person struct {
	Name string
}

func changeName(*person Person) {
	person.Name = "Changed name from function"
	fmt.Println(person)
}

func main() {
	person1 := Person{"Galih"}
	person2 := &person1
	person3 := Person{"Adhi"}

	person2.Name = "Kusuma"

	//fmt.Println(person1)
	//fmt.Println(person2)
	//fmt.Println(person3)

	person2 = &person3
	*person2 = Person{"Adhi 3"}

	//fmt.Println(person1)
	fmt.Println(person2)
	//fmt.Println(person3)

	changeName(person1)
}
