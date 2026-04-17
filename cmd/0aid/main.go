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
	case "module-plan":
		fs := flag.NewFlagSet("module-plan", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		plan := project.MilestoneModulePlan(bundle)
		if *out == "" {
			return printJSON(plan)
		}
		return project.WriteJSON(filepath.Clean(*out), plan)
	case "identity-plan":
		fs := flag.NewFlagSet("identity-plan", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		plan, err := project.IdentityFoundationPlan(bundle)
		if err != nil {
			return err
		}
		if *out == "" {
			return printJSON(plan)
		}
		return project.WriteJSON(filepath.Clean(*out), plan)
	case "signer-manifest":
		fs := flag.NewFlagSet("signer-manifest", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		manifest, err := project.SignerManifest(bundle)
		if err != nil {
			return err
		}
		if *out == "" {
			return printJSON(manifest)
		}
		return project.WriteJSON(filepath.Clean(*out), manifest)
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
	case "collect-validator":
		fs := flag.NewFlagSet("collect-validator", flag.ContinueOnError)
		bundlePath := fs.String("bundle", "", "node bundle directory")
		out := fs.String("out", "", "collected manifest output path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *bundlePath == "" {
			return errors.New("collect-validator requires --bundle")
		}
		if *out == "" {
			return errors.New("collect-validator requires --out")
		}
		manifest, err := project.CollectValidatorBundle(filepath.Clean(*bundlePath))
		if err != nil {
			return err
		}
		return project.WriteCollectedValidator(filepath.Clean(*out), manifest)
	case "assemble-genesis":
		fs := flag.NewFlagSet("assemble-genesis", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		collection := fs.String("collection", "", "directory containing collected validator manifests")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *collection == "" {
			return errors.New("assemble-genesis requires --collection")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		manifests, err := project.LoadCollectedValidators(filepath.Clean(*collection))
		if err != nil {
			return err
		}
		plan, err := project.AssembleGenesisPlan(bundle, manifests)
		if err != nil {
			return err
		}
		if *out == "" {
			return printJSON(plan)
		}
		return project.WriteJSON(filepath.Clean(*out), plan)
	case "assemble-localnet":
		fs := flag.NewFlagSet("assemble-localnet", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		collection := fs.String("collection", "", "directory containing collected validator manifests")
		out := fs.String("out", "", "output directory")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *collection == "" {
			return errors.New("assemble-localnet requires --collection")
		}
		if *out == "" {
			return errors.New("assemble-localnet requires --out")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		manifests, err := project.LoadCollectedValidators(filepath.Clean(*collection))
		if err != nil {
			return err
		}
		localnetBundle, err := project.RenderCollectedLocalnetBundle(bundle, manifests)
		if err != nil {
			return err
		}
		return project.WriteCollectedLocalnetBundle(filepath.Clean(*out), localnetBundle)
	default:
		return usageError()
	}
}

func usageError() error {
	return errors.New(
		"usage: 0aid <version|module-map|module-plan|identity-plan|signer-manifest|show-plan|init-genesis|render-validator|render-identity|init-node|collect-validator|assemble-genesis|assemble-localnet> [flags]",
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
