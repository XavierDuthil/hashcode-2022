package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// const exampleFile = "a_an_example"

func main() {
	files, err := ioutil.ReadDir("inputs")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		example := strings.Split(file.Name(), ".")[0]
		parseFile(example)
	}
}

func parseFile(exampleFile string) {
	content, err := ioutil.ReadFile(fmt.Sprintf("inputs/%s.in.txt", exampleFile))
	if err != nil {
		log.Fatal(err)
	}

	ingredientsLiked := make(map[string]int)
	ingredientsDisliked := make(map[string]int)

	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	// Parse first line
	scanner.Scan()
	fmt.Printf("%s clients to satisfy\n\n", scanner.Text())

	// Parse next lines
	i := 0
	for scanner.Scan() {
		line := scanner.Text()

		ingredients := parseLine(line)
		for _, ingredient := range ingredients {
			if i%2 == 0 {
				addIngredientToMap(ingredientsLiked, ingredient)
			} else {
				addIngredientToMap(ingredientsDisliked, ingredient)
			}
		}
		i++
	}

	ingredientsNotDisliked := getIngredientsNotDIsliked(ingredientsLiked, ingredientsDisliked)

	result := fmt.Sprintf("%d %s", len(ingredientsNotDisliked), strings.Join(ingredientsNotDisliked, " "))

	err = os.WriteFile(fmt.Sprintf("outputs/%s.out.txt", exampleFile), []byte(result), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ingredientsLiked)
	fmt.Println(ingredientsDisliked)
}

func parseLine(line string) []string {
	return RemoveIndex(strings.Fields(line), 0)
}

func RemoveIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func addIngredientToMap(ingredientsMap map[string]int, ingredient string) map[string]int {
	if _, ok := ingredientsMap[ingredient]; ok {
		ingredientsMap[ingredient] = ingredientsMap[ingredient] + 1
	} else {
		ingredientsMap[ingredient] = 1
	}

	return ingredientsMap
}

func getIngredientsNotDIsliked(ingredientsLiked map[string]int, ingredientsDisliked map[string]int) []string {
	var ingredientsNotDisliked []string

OUTER:
	for ingredientLiked, _ := range ingredientsLiked {
		for ingredientDisliked, _ := range ingredientsDisliked {
			if ingredientDisliked == ingredientLiked {
				continue OUTER
			}
		}
		ingredientsNotDisliked = append(ingredientsNotDisliked, ingredientLiked)
	}

	return ingredientsNotDisliked
}
