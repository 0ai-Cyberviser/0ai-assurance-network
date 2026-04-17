package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0ai-Cyberviser/0ai-assurance-network/internal/project"
)

const version = "0.1.0-dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "version":
		fmt.Printf("0aid %s\n", version)
		return nil
	case "module-map":
		return printJSON(project.ModuleMap())
	case "show-plan":
		fs := flag.NewFlagSet("show-plan", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		return printJSON(project.NetworkPlan(bundle))
	case "init-genesis":
		fs := flag.NewFlagSet("init-genesis", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		rendered := project.RenderedGenesis(bundle)
		if *out == "" {
			return printJSON(rendered)
		}
		target := filepath.Clean(*out)
		return project.WriteJSON(target, rendered)
	case "render-validator":
		fs := flag.NewFlagSet("render-validator", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		validatorID := fs.String("id", "", "validator id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *validatorID == "" {
			return errors.New("render-validator requires --id")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		plan, err := project.ValidatorPlan(bundle, *validatorID)
		if err != nil {
			return err
		}
		return printJSON(plan)
	case "render-identity":
		fs := flag.NewFlagSet("render-identity", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		validatorID := fs.String("id", "", "validator id")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *validatorID == "" {
			return errors.New("render-identity requires --id")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		identity, err := project.DevelopmentIdentityForValidator(bundle, *validatorID)
		if err != nil {
			return err
		}
		return printJSON(identity)
	case "init-node":
		fs := flag.NewFlagSet("init-node", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		validatorID := fs.String("id", "", "validator id")
		out := fs.String("out", "", "output directory")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *validatorID == "" {
			return errors.New("init-node requires --id")
		}
		if *out == "" {
			return errors.New("init-node requires --out")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		nodeBundle, err := project.NodeInitPlan(bundle, *validatorID)
		if err != nil {
			return err
		}
		return project.WriteNodeInitBundle(filepath.Clean(*out), nodeBundle)
	default:
		return usageError()
	}
}

func usageError() error {
	return errors.New(
		"usage: 0aid <version|module-map|show-plan|init-genesis|render-validator|render-identity|init-node> [flags]",
	)
}

func printJSON(value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", encoded)
	return nil
}
