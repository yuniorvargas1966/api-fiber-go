package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

var db *sql.DB

// ================= DATABASE =================
func initDB() {
	if err := godotenv.Load(); err != nil {
		log.Println("No se pudo cargar .env")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("Usuario"),
		os.Getenv("Contrasena"),
		os.Getenv("Host"),
		os.Getenv("PortDB"),
		os.Getenv("Nombre"),
	)

	var err error
	db, err = sql.Open(os.Getenv("Driver"), dsn)
	if err != nil {
		log.Fatal("Error conectando a DB:", err)
	}

	// Pool de conexiones (CLAVE)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal("DB no responde:", err)
	}

	log.Println("✅ Conectado a MySQL")
}

// ================= MODELO =================
type Servicio struct {
	ID          int    `json:"id"`
	Nombre      string `json:"nombre"`
	Correo      string `json:"correo"`
	Telefono    string `json:"telefono"`
	Equipo      string `json:"equipo"`
	Diagnostico string `json:"diagnostico"`
	Resultados  string `json:"resultados"`
	Decision    string `json:"decision"`
	Taller      string `json:"taller"`
	Servicio    string `json:"servicio"`
	Entrega     string `json:"entrega"`
	Fecha       string `json:"fecha"`
}

// ================= HANDLERS =================
func getServicios(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
		SELECT id, nombre, correo, telefono, equipo, diagnostico, resultados, decision, taller, servicio, entrega, fecha 
		FROM taller`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var servicios []Servicio

	for rows.Next() {
		var s Servicio
		if err := rows.Scan(
			&s.ID, &s.Nombre, &s.Correo, &s.Telefono,
			&s.Equipo, &s.Diagnostico, &s.Resultados,
			&s.Decision, &s.Taller, &s.Servicio,
			&s.Entrega, &s.Fecha,
		); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		servicios = append(servicios, s)
	}

	return c.JSON(servicios)
}

func getServicio(c *fiber.Ctx) error {
	id := c.Params("id")

	var s Servicio

	err := db.QueryRow(`
		SELECT id, nombre, correo, telefono, equipo, diagnostico, resultados, decision, taller, servicio, entrega, fecha 
		FROM taller WHERE id=?`, id).
		Scan(
			&s.ID, &s.Nombre, &s.Correo, &s.Telefono,
			&s.Equipo, &s.Diagnostico, &s.Resultados,
			&s.Decision, &s.Taller, &s.Servicio,
			&s.Entrega, &s.Fecha,
		)

	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "No encontrado"})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(s)
}

func createServicio(c *fiber.Ctx) error {
	var s Servicio

	if err := c.BodyParser(&s); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	res, err := db.Exec(`
		INSERT INTO taller 
		(nombre, correo, telefono, equipo, diagnostico, resultados, decision, taller, servicio, entrega, fecha)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.Nombre, s.Correo, s.Telefono, s.Equipo,
		s.Diagnostico, s.Resultados, s.Decision,
		s.Taller, s.Servicio, s.Entrega, s.Fecha,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := res.LastInsertId()
	s.ID = int(id)

	return c.Status(201).JSON(s)
}

func updateServicio(c *fiber.Ctx) error {
	id := c.Params("id")
	var s Servicio

	if err := c.BodyParser(&s); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON inválido"})
	}

	result, err := db.Exec(`
		UPDATE taller SET 
		nombre=?, correo=?, telefono=?, equipo=?, diagnostico=?, resultados=?, 
		decision=?, taller=?, servicio=?, entrega=?, fecha=? 
		WHERE id=?`,
		s.Nombre, s.Correo, s.Telefono, s.Equipo,
		s.Diagnostico, s.Resultados, s.Decision,
		s.Taller, s.Servicio, s.Entrega, s.Fecha, id,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "No encontrado"})
	}

	return c.JSON(fiber.Map{"message": "Actualizado correctamente"})
}

func deleteServicio(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := db.Exec("DELETE FROM taller WHERE id=?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "No encontrado"})
	}

	return c.JSON(fiber.Map{"message": "Eliminado correctamente"})
}

// ================= MAIN =================
func main() {
	initDB()
	defer db.Close()

	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	// Rutas
	api := app.Group("/api")

	api.Get("/servicios", getServicios)
	api.Get("/servicios/:id", getServicio)
	api.Post("/servicios", createServicio)
	api.Put("/servicios/:id", updateServicio)
	api.Delete("/servicios/:id", deleteServicio)

	port := os.Getenv("PORT")
	if port == "" {
		port = "Port"
	}

	log.Fatal(app.Listen("localhost" + port))
}
