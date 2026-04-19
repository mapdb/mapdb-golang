// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

//go:build kata

package object

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// Pet Kata — Eclipse Collections for Go
//
// This is a learning exercise. Each test has a TODO where you need to
// write code using the object collection API. The assertions are already
// written — your job is to make them pass.
//
// Domain model (provided): Person, Pet, PetType
// Collections to use: ArrayList[T], HashSet[T], HashBag[T], HashMap[K,V]
//
// Run:  go test -run PetKata -v
// ═══════════════════════════════════════════════════════════════════════

// PetType represents a type of pet (used as enum-like values).
type PetType int

const (
	CAT PetType = iota
	DOG
	HAMSTER
	TURTLE
	BIRD
	SNAKE
)

func (p PetType) String() string {
	switch p {
	case CAT:
		return "CAT"
	case DOG:
		return "DOG"
	case HAMSTER:
		return "HAMSTER"
	case TURTLE:
		return "TURTLE"
	case BIRD:
		return "BIRD"
	case SNAKE:
		return "SNAKE"
	default:
		return "UNKNOWN"
	}
}

// Pet represents a pet with a name and type.
type Pet struct {
	Name    string
	PetType PetType
}

// Person represents a person who may own pets.
type Person struct {
	FirstName string
	LastName  string
	Pets      *ArrayList[Pet]
}

func (p Person) HasPetType(pt PetType) bool {
	return p.Pets.AnySatisfy(func(pet Pet) bool { return pet.PetType == pt })
}

func (p Person) FullName() string {
	return p.FirstName + " " + p.LastName
}

func setupPeople() *ArrayList[Person] {
	people := NewArrayList[Person]()

	mary := Person{FirstName: "Mary", LastName: "Smith", Pets: NewArrayListFrom(
		Pet{Name: "Tabby", PetType: CAT},
	)}
	bob := Person{FirstName: "Bob", LastName: "Smith", Pets: NewArrayListFrom(
		Pet{Name: "Dolly", PetType: DOG},
		Pet{Name: "Spot", PetType: DOG},
	)}
	ted := Person{FirstName: "Ted", LastName: "Smith", Pets: NewArrayListFrom(
		Pet{Name: "Spike", PetType: DOG},
		Pet{Name: "Serpy", PetType: SNAKE},
	)}
	jake := Person{FirstName: "Jake", LastName: "Snake", Pets: NewArrayListFrom(
		Pet{Name: "Speedy", PetType: TURTLE},
		Pet{Name: "Tweety", PetType: BIRD},
	)}
	barry := Person{FirstName: "Barry", LastName: "Jones", Pets: NewArrayListFrom(
		Pet{Name: "Fluffy", PetType: CAT},
		Pet{Name: "Crunchie", PetType: HAMSTER},
	)}
	terry := Person{FirstName: "Terry", LastName: "Schneider", Pets: NewArrayListFrom(
		Pet{Name: "Cozy", PetType: CAT},
		Pet{Name: "Rumple", PetType: HAMSTER},
	)}
	harry := Person{FirstName: "Harry", LastName: "Harrison", Pets: NewArrayList[Pet]()}

	people.Add(mary)
	people.Add(bob)
	people.Add(ted)
	people.Add(jake)
	people.Add(barry)
	people.Add(terry)
	people.Add(harry)

	return people
}

// Exercise 1: Do any people have cats?
// Hint: Use AnySatisfy with the HasPetType helper.
func TestPetKata_DoAnyPeopleHaveCats(t *testing.T) {
	people := setupPeople()
	_ = people // TODO: replace false with an expression using people.AnySatisfy(...)
	hasCats := false

	if !hasCats {
		t.Error("expected someone to have a cat")
	}
}

// Exercise 2: Do all people have pets?
// Hint: Use AllSatisfy. Harry has no pets, so this should be false.
func TestPetKata_DoAllPeopleHavePets(t *testing.T) {
	people := setupPeople()
	_ = people // TODO: replace true with an expression using people.AllSatisfy(...)
	allHavePets := true

	if allHavePets {
		t.Error("expected not all people to have pets (Harry has none)")
	}
}

// Exercise 3: Does nobody have snakes?
// Hint: Use NoneSatisfy. Ted has a snake, so this should be false.
func TestPetKata_DoesNobodyHaveSnakes(t *testing.T) {
	people := setupPeople()
	_ = people // TODO: replace true with an expression using people.NoneSatisfy(...)
	noSnakes := true

	if noSnakes {
		t.Error("expected NoneSatisfy(snake) = false (Ted has a snake)")
	}
}

// Exercise 4: How many people have cats?
// Hint: Use Count with a predicate.
func TestPetKata_HowManyPeopleHaveCats(t *testing.T) {
	people := setupPeople()
	_ = people // TODO: replace 0 with people.Count(...)
	catPeopleCount := 0

	if catPeopleCount != 3 {
		t.Errorf("Count(hasCat) = %d, want 3", catPeopleCount)
	}
}

// Exercise 5: Get the people who have cats.
// Hint: Use Select to filter the list.
func TestPetKata_GetPeopleWithCats(t *testing.T) {
	people := setupPeople()
	_ = people // TODO: replace with people.Select(...)
	catPeople := NewArrayList[Person]()

	if catPeople.Size() != 3 {
		t.Errorf("Select(hasCat) size = %d, want 3", catPeople.Size())
	}
	names := NewHashSet[string]()
	catPeople.ForEach(func(p Person) { names.Add(p.FirstName) })
	for _, name := range []string{"Mary", "Barry", "Terry"} {
		if !names.Contains(name) {
			t.Errorf("expected %s in cat people", name)
		}
	}
}

// Exercise 6: Get the people who do NOT have cats.
// Hint: Use Reject.
func TestPetKata_GetPeopleWithoutCats(t *testing.T) {
	people := setupPeople()
	_ = people // TODO: replace with people.Reject(...)
	noCatPeople := NewArrayList[Person]()

	if noCatPeople.Size() != 4 {
		t.Errorf("Reject(hasCat) size = %d, want 4", noCatPeople.Size())
	}
}

// Exercise 7: Find Mary Smith.
// Hint: Use Detect with a predicate that checks FirstName and LastName.
func TestPetKata_FindMarySmith(t *testing.T) {
	people := setupPeople()
	_ = people // TODO: use people.Detect(...)
	var mary Person
	found := false

	if !found {
		t.Fatal("Detect did not find Mary Smith")
	}
	if mary.FullName() != "Mary Smith" {
		t.Errorf("Detect found %q, want 'Mary Smith'", mary.FullName())
	}
	if mary.Pets.Size() != 1 {
		t.Errorf("Mary has %d pets, want 1", mary.Pets.Size())
	}
}

// Exercise 8: Collect all pet names across all people.
// Hint: Use ForEach on people, then ForEach on each person's pets.
func TestPetKata_CollectAllPetNames(t *testing.T) {
	people := setupPeople()
	petNames := NewArrayList[string]()
	_ = people // TODO: iterate people and their pets, adding each pet name to petNames

	if petNames.Size() != 11 {
		t.Errorf("total pet names = %d, want 11", petNames.Size())
	}
	if !petNames.Contains("Tabby") {
		t.Error("missing pet name Tabby")
	}
	if !petNames.Contains("Tweety") {
		t.Error("missing pet name Tweety")
	}
}

// Exercise 9: Count pet types using a HashBag.
// Hint: Create a HashBag[PetType], iterate all people's pets, Add each pet's type.
func TestPetKata_CountPetTypes(t *testing.T) {
	people := setupPeople()
	petTypeBag := NewHashBag[PetType]()
	_ = people // TODO: populate petTypeBag

	if petTypeBag.OccurrencesOf(CAT) != 3 {
		t.Errorf("CAT occurrences = %d, want 3", petTypeBag.OccurrencesOf(CAT))
	}
	if petTypeBag.OccurrencesOf(DOG) != 3 {
		t.Errorf("DOG occurrences = %d, want 3", petTypeBag.OccurrencesOf(DOG))
	}
	if petTypeBag.OccurrencesOf(HAMSTER) != 2 {
		t.Errorf("HAMSTER occurrences = %d, want 2", petTypeBag.OccurrencesOf(HAMSTER))
	}
	if petTypeBag.Size() != 11 {
		t.Errorf("total pet type bag size = %d, want 11", petTypeBag.Size())
	}

	top := petTypeBag.TopOccurrences(2)
	if len(top) != 2 {
		t.Fatalf("TopOccurrences(2) len = %d, want 2", len(top))
	}
	if top[0].Count != 3 {
		t.Errorf("top[0].Count = %d, want 3", top[0].Count)
	}
}

// Exercise 10: Collect all unique pet types using a HashSet.
// Hint: Create a HashSet[PetType], iterate all pets, Add each type.
func TestPetKata_UniquePetTypes(t *testing.T) {
	people := setupPeople()
	petTypes := NewHashSet[PetType]()
	_ = people // TODO: populate petTypes

	if petTypes.Size() != 6 {
		t.Errorf("unique pet types = %d, want 6", petTypes.Size())
	}
	for _, pt := range []PetType{CAT, DOG, HAMSTER, TURTLE, BIRD, SNAKE} {
		if !petTypes.Contains(pt) {
			t.Errorf("missing pet type %v", pt)
		}
	}
}

// Exercise 11: Count total pets using InjectInto (fold).
// Hint: Use people.InjectInto(0, func(acc any, p Person) any { ... })
func TestPetKata_TotalPetCount(t *testing.T) {
	people := setupPeople()
	_ = people // TODO: replace 0 with people.InjectInto(...)
	total := 0

	if total != 11 {
		t.Errorf("total pets via InjectInto = %d, want 11", total)
	}
}

// Exercise 12: Detect a person who doesn't exist.
// Hint: Detect should return (zero, false) when no match is found.
func TestPetKata_DetectNotFound(t *testing.T) {
	people := setupPeople()
	_ = people // TODO: use people.Detect(...) to search for "Nobody Here"
	found := true

	if found {
		t.Error("Detect found a person that should not exist")
	}
}
