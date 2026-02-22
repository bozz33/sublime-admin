package commands

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bozz33/sublimeadmin/auth"
	"github.com/bozz33/sublimeadmin/engine"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"s"},
	Short:   "Start the HTTP server",
	Long: `Start the SublimeGo HTTP server with graceful shutdown.
	
This command:
- Loads the configuration
- Connects to the database
- Initializes the admin panel
- Starts the HTTP server
- Handles graceful shutdown on SIGINT/SIGTERM`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()

		fmt.Printf("Démarrage de SublimeGo %s\n", version)
		fmt.Printf("Environnement: %s\n", cfg.Environment)
		fmt.Printf("Mode debug: %v\n", cfg.App.Debug)

		// Connexion à la base de données
		fmt.Printf("Connexion à la base de données (%s)...\n", cfg.Database.Driver)

		// Ajouter le pragma foreign_keys pour SQLite si absent
		dbURL := cfg.Database.URL
		if (cfg.Database.Driver == "sqlite3" || cfg.Database.Driver == "sqlite") && !strings.Contains(dbURL, "_fk=") {
			if strings.Contains(dbURL, "?") {
				dbURL += "&_fk=1"
			} else {
				dbURL += "?_fk=1"
			}
		}
		fmt.Printf("URL: %s\n", dbURL)

		// Ouvrir la connexion SQL
		db, err := sql.Open(cfg.Database.Driver, dbURL)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Activer les foreign keys pour SQLite
		if cfg.Database.Driver == "sqlite3" || cfg.Database.Driver == "sqlite" {
			if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
				return fmt.Errorf("failed to enable foreign keys: %w", err)
			}
		}

		if err := db.Ping(); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}
		fmt.Println("✓ Connexion DB réussie")

		// Initialisation de la session et de l'authentification
		fmt.Println("Initialisation de l'authentification...")
		sessionManager := scs.New()
		authManager := auth.NewManager(sessionManager)

		// Initialisation du panneau d'administration
		// Note: pass your ORM client via panel.WithUsers(yourUserRepo) in your project
		fmt.Println("Initialisation du panneau d'administration...")
		panel := engine.NewPanel("admin").
			WithPath(cfg.Engine.BasePath).
			WithBrandName(cfg.Engine.BrandName).
			WithAuthManager(authManager).
			WithSession(sessionManager)

		// Création du serveur HTTP
		addr := cfg.ServerAddress()
		server := &http.Server{
			Addr:              addr,
			Handler:           panel.Router(),
			ReadTimeout:       cfg.Server.ReadTimeout,
			WriteTimeout:      cfg.Server.WriteTimeout,
			IdleTimeout:       cfg.Server.IdleTimeout,
			MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
			ReadHeaderTimeout: 10 * time.Second,
		}

		// Démarrage du serveur en goroutine
		serverErrors := make(chan error, 1)
		go func() {
			fmt.Printf("\nServeur démarré sur http://%s\n", addr)
			fmt.Printf("Panneau d'administration: http://%s%s\n", addr, cfg.Engine.BasePath)
			fmt.Println("\nAppuyez sur Ctrl+C pour arrêter")
			serverErrors <- server.ListenAndServe()
		}()

		// Attente du signal d'interruption
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		select {
		case err := <-serverErrors:
			return fmt.Errorf("server error: %w", err)
		case sig := <-shutdown:
			fmt.Printf("\n\nSignal reçu: %v\n", sig)
			fmt.Println("Arrêt gracieux en cours...")

			// Arrêt gracieux du serveur
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				server.Close()
				return fmt.Errorf("failed to shutdown gracefully: %w", err)
			}

			fmt.Println("Serveur arrêté avec succès")
		}

		return nil
	},
}
