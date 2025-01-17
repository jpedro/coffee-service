package data

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-memdb"

	"github.com/hashicorp-demoapp/coffee-service/config"
	"github.com/hashicorp-demoapp/coffee-service/data/entities"
)

// TableNameKey is a typesafe discriminator for table names
type TableNameKey string

func (t TableNameKey) String() string {
	return string(t)
}

const (
	// Ingredient is the ingredient table name
	Ingredient TableNameKey = "ingredient"
	// Coffee is the coffee table name
	Coffee TableNameKey = "coffee"
	// CoffeeIngredient is the coffee_ingredient table name
	CoffeeIngredient TableNameKey = "coffee_ingredient"
)

// InMemoryRepository implements the coffee-service.data.Repository interface
// uisng go-membdb instead of postgres.
type InMemoryRepository struct {
	db     *memdb.MemDB
	config *config.Config
}

// NewInMemoryDB is the InMemoryRepository factory method. It fulfills the same
// interface as Repository, but uses go-membdb internally to provide data. NOTE,
// this interface requires build time tooling.
func NewInMemoryDB(config *config.Config) (Repository, error) {
	config.Logger.Debug("Attempting to load in memory db")
	// Create a new data base
	db, err := memdb.NewMemDB(createSchema())
	if err != nil {
		config.Logger.Debug(fmt.Sprintf("Failed to load in membory database with err %+v", err))
		return &InMemoryRepository{}, err
	}

	repository := &InMemoryRepository{db, config}

	repository.config.Logger.Debug("Loading Ingredients")
	err = repository.loadIngredients()
	if err != nil {
		repository.config.Logger.Debug(fmt.Sprintf("Failed to load ingredients with err %+v", err))
		return &InMemoryRepository{}, err
	}

	repository.config.Logger.Debug("Loading coffees")
	err = repository.loadCoffees()
	if err != nil {
		repository.config.Logger.Debug(fmt.Sprintf("Failed to load coffees with err %+v", err))
		return &InMemoryRepository{}, err
	}

	repository.config.Logger.Debug("Loading coffee ingredients")
	err = repository.loadCoffeeIngredients()
	if err != nil {
		repository.config.Logger.Debug(fmt.Sprintf("Failed to load coffee ingredients with err %+v", err))
		return &InMemoryRepository{}, err
	}

	repository.config.Logger.Debug("Data loaded")
	return repository, nil
}

// Find returns all coffees from the database
// Used to accept ctx opentracing.SpanContext
func (r *InMemoryRepository) Find() (entities.Coffees, error) {
	txn := r.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(Coffee.String(), "id")
	if err != nil {
		r.config.Logger.Error("coffee-service.data.InMemoryRepository.Find failed to load coffees", err)
		return nil, err
	}

	coffees := make([]entities.Coffee, 0)

	for coffee := iter.Next(); coffee != nil; coffee = iter.Next() {
		coffees = append(coffees, *coffee.(*entities.Coffee))
	}

	for _, coffee := range coffees {
		coffeeIngredients := make([]entities.CoffeeIngredients, 0)

		innerIter, err := txn.Get(CoffeeIngredient.String(), "id")
		if err != nil {
			r.config.Logger.Error("coffee-service.data.InMemoryRepository.Find failed to load ingredients", err)
			return nil, err
		}

		for ingredient := innerIter.Next(); ingredient != nil; ingredient = innerIter.Next() {
			coffeeIngredients = append(coffeeIngredients, *ingredient.(*entities.CoffeeIngredients))
			fmt.Printf("coffee-service.data.InMemoryRepository.Find loaded ingredients %s\n", coffeeIngredients)
		}

		coffee.Ingredients = coffeeIngredients
	}

	return coffees, nil
}

func createSchema() *memdb.DBSchema {
	// Create the DB schema
	// TODO Update to this entities with tooling.
	return &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			Coffee.String(): {
				Name: Coffee.String(),
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "ID"},
					},
				},
			},
			Ingredient.String(): {
				Name: Ingredient.String(),
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "ID"},
					},
				},
			},
			CoffeeIngredient.String(): {
				Name: CoffeeIngredient.String(),
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "ID"},
					},
				},
			},
		},
	}
}

func (r *InMemoryRepository) loadIngredients() error {
	timestamp := time.Now().String()
	txn := r.db.Txn(true)

	// Insert some people
	ingredients := []*entities.Ingredient{
		{ID: 1, Name: "Espresso'", CreatedAt: timestamp, UpdatedAt: timestamp},
		{ID: 2, Name: "Semi Skimmed Milk", CreatedAt: timestamp, UpdatedAt: timestamp},
		{ID: 3, Name: "Hot Water", CreatedAt: timestamp, UpdatedAt: timestamp},
		{ID: 4, Name: "Pumpkin Spice", CreatedAt: timestamp, UpdatedAt: timestamp},
		{ID: 5, Name: "Steamed Milk", CreatedAt: timestamp, UpdatedAt: timestamp},
	}

	for _, row := range ingredients {
		if err := txn.Insert(Ingredient.String(), row); err != nil {
			return err
		}
		fmt.Printf("Loaded ingredient %+v\n", row)
	}

	txn.Commit()
	return nil
}

func (r *InMemoryRepository) loadCoffees() error {
	timestamp := time.Now().String()
	txn := r.db.Txn(true)

	coffees := []*entities.Coffee{
		{
			ID:          1,
			Name:        "Packer Spiced Latte",
			Teaser:      "Packed with goodness to spice up your images",
			Description: "",
			Price:       350,
			Image:       "/packer.png",
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
		},
		{
			ID:          2,
			Name:        "Vaulatte",
			Teaser:      "Nothing gives you a safe and secure feeling like a Vaulatte",
			Description: "",
			Price:       200,
			Image:       "/vault.png",
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
		},
		{
			ID:          3,
			Name:        "Nomadicano",
			Teaser:      "Drink one today and you will want to schedule another",
			Description: "",
			Price:       150,
			Image:       "/nomad.png",
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
		},
		{
			ID:          4,
			Name:        "Terraspresso",
			Teaser:      "Nothing kickstarts your day like a provision of Terraspresso",
			Description: "",
			Price:       150,
			Image:       "/terraform.png",
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
		},
		{
			ID:          5,
			Name:        "Vagrante espresso",
			Teaser:      "Stdin is not a tty",
			Description: "",
			Price:       200,
			Image:       "/vagrant.png",
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
		},
		{
			ID:          6,
			Name:        "Connectaccino",
			Teaser:      "Discover the wonders of our meshy service",
			Description: "",
			Price:       250,
			Image:       "/consul.png",
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
		},
	}

	for _, c := range coffees {
		if err := txn.Insert(Coffee.String(), c); err != nil {
			return err
		}
	}

	txn.Commit()
	return nil
}

func (r *InMemoryRepository) loadCoffeeIngredients() error {
	timestamp := time.Now().String()
	txn := r.db.Txn(true)

	coffeeIngredients := []*entities.CoffeeIngredients{
		{
			ID:           1,
			CoffeeID:     1,
			IngredientID: 1,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
		{
			ID:           2,
			CoffeeID:     1,
			IngredientID: 2,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
		{
			ID:           3,
			CoffeeID:     1,
			IngredientID: 4,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
		{
			ID:           4,
			CoffeeID:     2,
			IngredientID: 1,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
		{
			ID:           5,
			CoffeeID:     2,
			IngredientID: 2,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
		{
			ID:           6,
			CoffeeID:     3,
			IngredientID: 1,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
		{
			ID:           7,
			CoffeeID:     3,
			IngredientID: 3,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
		{
			ID:           8,
			CoffeeID:     4,
			IngredientID: 1,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
		{
			ID:           9,
			CoffeeID:     5,
			IngredientID: 1,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
		{
			ID:           10,
			CoffeeID:     6,
			IngredientID: 1,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
		{
			ID:           11,
			CoffeeID:     6,
			IngredientID: 5,
			CreatedAt:    timestamp,
			UpdatedAt:    timestamp,
		},
	}

	for _, ci := range coffeeIngredients {
		if err := txn.Insert(CoffeeIngredient.String(), ci); err != nil {
			return err
		}
	}

	txn.Commit()
	return nil
}
