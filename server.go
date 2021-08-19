package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// DB configuration
	dbUrl := "<url>"

	client, err := mongo.NewClient(options.Client().ApplyURI(dbUrl))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	database := client.Database("teamvotedb")
	usersCollection := database.Collection("users_test")

	// Server configuration
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, world!")
	})

	e.GET("/users", func(c echo.Context) error {
		cursor, err := usersCollection.Find(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
			return err
		}

		var users []User
		for cursor.Next(ctx) {
			var user User
			err := cursor.Decode(&user)
			if err != nil {
				log.Fatal(err)
				return err
			}
			users = append(users, user)
		}

		return c.JSON(http.StatusOK, users)
	})

	e.GET("/users/:id", func(c echo.Context) error {
		var user User

		id, _ := primitive.ObjectIDFromHex(c.Param(("id")))
		filter := bson.M{"_id": id}

		err := usersCollection.FindOne(ctx, filter).Decode(&user)
		if err != nil {
			log.Fatal(err)
			return err
		}

		return c.JSON(http.StatusCreated, user)
	})

	e.POST("/users", func(c echo.Context) error {
		u := new(User)
		if err := c.Bind(u); err != nil {
			log.Fatal(err)
			return err
		}

		insertResult, err := usersCollection.InsertOne(ctx, bson.D{
			{Key: "name", Value: u.Name},
			{Key: "email", Value: u.Email},
		})
		if err != nil {
			log.Fatal(err)
			return err
		}

		if oid, ok := insertResult.InsertedID.(primitive.ObjectID); ok {
			u.Id = oid
		}
		fmt.Printf("Iserted id: %s", insertResult.InsertedID)

		return c.JSON(http.StatusCreated, u)
	})

	e.PUT("/users/:id", func(c echo.Context) error {
		updateUser := new(User)
		if err := c.Bind(updateUser); err != nil {
			log.Fatal(err)
			return err
		}

		var existingUser User
		id, _ := primitive.ObjectIDFromHex(c.Param(("id")))
		filter := bson.M{"_id": id}
		err := usersCollection.FindOne(ctx, filter).Decode(&existingUser)
		if err != nil {
			log.Fatal(err)
			return err
		}

		update := bson.D{
			{
				Key: "$set", Value: bson.D{
					{Key: "name", Value: updateUser.Name},
					{Key: "email", Value: updateUser.Email},
				},
			},
		}

		var updatedUser User
		err = usersCollection.FindOneAndUpdate(ctx, filter, update).Decode(&updatedUser)
		if err != nil {
			log.Fatal(err)
			return err
		}
		return c.JSON(http.StatusCreated, updatedUser)
	})

	e.DELETE("/users/:id", func(c echo.Context) error {
		id, err := primitive.ObjectIDFromHex(c.Param(("id")))
		if err != nil {
			log.Fatal(err)
			return err
		}

		filter := bson.M{"_id": id}

		_, err = usersCollection.DeleteOne(ctx, filter)
		if err != nil {
			log.Fatal(err)
			return err
		}

		return c.JSON(http.StatusNoContent, "")
	})

	e.Logger.Fatal(e.Start(":8080"))
}

type User struct {
	Id    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name,omitempty" bson:"name,omitempty"`
	Email string             `json:"email,omitempty" bson:"email,omitempty"`
}
