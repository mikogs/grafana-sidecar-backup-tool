package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	gocli "github.com/mikogs/lib-go-cli"
	"os"
	"time"
	"log"
)

// DEFAULT_SLEEP is used when reading config fails and the program is set to ignore the error
const DEFAULT_SLEEP = 10

func main() {
	cli := gocli.NewCLI("grafana-sidecar-backup-tool", "", "Mikolaj Gasior <nl@gen64.net>")
	cmdStart := cli.AddCmd("start", "Starts the daemon", startHandler)
	cmdStart.AddFlag("config", "c", "config", "YAML file with users", gocli.TypePathFile|gocli.MustExist|gocli.Required, nil)
	cmdStart.AddFlag("quiet", "q", "", "Quite mode. Do not output anything", gocli.TypeBool, nil)
	cmdStart.AddFlag("ignore_errors", "i", "", "Ignore errors and continue", gocli.TypeBool, nil)
	_ = cli.AddCmd("version", "Prints version", versionHandler)
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		os.Args = []string{"App", "version"}
	}
	os.Exit(cli.Run(os.Stdout, os.Stderr))
}

func versionHandler(c *gocli.CLI) int {
	fmt.Fprintf(os.Stdout, VERSION+"\n")
	return 0
}

func readConfig(f string, cfg *Config) error {
	fmt.Fprintf(os.Stdout, "Reading config file %s...\n", f)
	err := cfg.SetFromYAMLFile(f)
	if err != nil {
		return fmt.Errorf("Error with config file: %w\n", err)
	}
	return nil
}

func connectToDB(cfg *Config) (*sql.DB, error) {
	if cfg.DryRun {
		fmt.Fprintf(os.Stderr, "Dry-running...\n")
		return nil, nil
	}
	db, err := sql.Open("sqlite3", cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("Error with connecting to the database: %w\n", err)
	}
	return db, err

}

func processDatasources(cfg *Config, db *sql.DB) error {
}

func processDashboards(cfg *Config, db *sql.DB) error {
	rows, err := db.Query("SELECT d.id, d.version, d.slug, d.title, d.folder_id, d.is_folder, df.slug AS folder_slug, df.title AS folder_title, df.folder_id AS folder_folder_id, df.is_folder AS folder_is_folder FROM dashboard d LEFT JOIN dashboard df ON d.folder_id = df.id;")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id int64
			version int64
			slug string
			title string
			folder_id int64
			is_folder int
			folder_slug string
			folder_title string
			folder_folder_id string
			org_id int64
		)
		if err := rows.Scan(&id, &title, &version, &org_id); err != nil {
			return err
		}
	}

	return nil
}

func processContactPoints(cfg *Config, db *sql.DB) error {
}

func processAlerts(cfg *Config, db *sql.DB) error {

}

/*func update(role string, login string, org int, db *sql.DB, cfg *Config) error {
	fmt.Fprintf(os.Stdout, "Setting login '%s' to %s for org %d...\n", login, role, org)

	if cfg.DryRun {
		return nil
	}

	if _, err := db.Exec("UPDATE org_user SET role = ? WHERE user_id IN (SELECT id FROM user WHERE login=?);", role, login); err != nil {
		return fmt.Errorf("UPDATE query for Viewer failed to execute: %w", err)
	}
	return nil

}

func updateOrgs(cfg *Config, db *sql.DB) error {
	for _, org := range cfg.Orgs {
		fmt.Fprintf(os.Stdout, "Got org %v from the config file\n", org.ID)
		for _, viewer := range org.Viewers {
			err := update("Viewer", viewer.Login, org.ID, db, cfg)
			if err != nil {
				return err
			}
		}
		for _, editor := range org.Editors {
			err := update("Editor", editor.Login, org.ID, db, cfg)
			if err != nil {
				return err
			}
		}
		for _, admin := range org.Admins {
			err := update("Admin", admin.Login, org.ID, db, cfg)
			if err != nil {
				return err
			}
		}
	}
	return nil

}*/

func startHandler(c *gocli.CLI) int {
	ch := make(chan int)
	go func(ch chan int) {
		var cfg Config
		var db *sql.DB
		for {
			// read env with github token
			err := readConfig(c.Flag("config"), &cfg)
			if err != nil {
				cfg.Sleep = DEFAULT_SLEEP
			}
			if err == nil {
				db, err = connectToDB(&cfg)
			}
			if err == nil {
				// checkout git somewhere
				err = processDatasources(&cfg, db)
				err = processDashboards(&cfg, db)
				err = processContactPoints(&cfg, db)
				err = processAlerts(&cfg, db)
				// process dashboards
				// process alerts
				// process contact points
				// ...
				// ---
				// select all dashboards and prepare JSON files for them
				// do a version with split JSON files
				// select all alerts (and contacts points) and generate YAML files for them
				// ---
				// checkout latest git repository (in some tmp directory that can be removed)
				// for the all-in-one json dashboard file - compare the latest ones (trimmed,minified)
				// write a comparison that includes panel comparison between jsons
				// OR should we just map the versions of dashboards 1-to-1?
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err.Error())
				if c.Flag("ignore_errors") == "false" || cfg.RunOnce {
					ch <- 1
				} else {
					fmt.Fprintf(os.Stderr, "Ignoring error and continuing to do nothing...\n")
				}
			}

			if !cfg.DryRun {
				db.Close()
			}
			if cfg.RunOnce {
				break
			} else {
				fmt.Fprintf(os.Stdout, "Sleeping %d seconds...\n", cfg.Sleep)
				time.Sleep(time.Duration(cfg.Sleep) * time.Second)
			}
		}
		ch <- 0
	}(ch)
	lastErr := <-ch
	return lastErr
}
