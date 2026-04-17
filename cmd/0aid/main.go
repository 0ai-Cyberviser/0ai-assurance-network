package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0ai-Cyberviser/0ai-assurance-network/internal/project"
)

const version = "0.1.0-dev"
const checkpointSignerPolicyPath = "config/governance/checkpoint-signers.json"

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
	case "signer-rotation-receipt":
		fs := flag.NewFlagSet("signer-rotation-receipt", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		outgoingSignerID := fs.String("outgoing-signer-id", "", "outgoing signer id")
		incomingSignerID := fs.String("incoming-signer-id", "", "incoming signer id")
		incomingKeyID := fs.String("incoming-key-id", "", "incoming key id")
		incomingActorID := fs.String("incoming-actor-id", "", "incoming actor id (defaults to outgoing actor)")
		incomingRoles := fs.String("incoming-roles", "", "comma-separated incoming roles (defaults to outgoing roles)")
		incomingProvisionedAt := fs.String("incoming-provisioned-at", "", "incoming signer provisioned_at timestamp")
		incomingRotateBy := fs.String("incoming-rotate-by", "", "incoming signer rotate_by timestamp")
		effectiveAt := fs.String("effective-at", "", "rotation effective_at timestamp")
		receiptID := fs.String("receipt-id", "", "explicit receipt id")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *outgoingSignerID == "" {
			return errors.New("signer-rotation-receipt requires --outgoing-signer-id")
		}
		if *incomingSignerID == "" {
			return errors.New("signer-rotation-receipt requires --incoming-signer-id")
		}
		if *incomingKeyID == "" {
			return errors.New("signer-rotation-receipt requires --incoming-key-id")
		}
		if *effectiveAt == "" {
			return errors.New("signer-rotation-receipt requires --effective-at")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		request := project.SignerRotationReceiptRequest{
			OutgoingSignerID:      *outgoingSignerID,
			IncomingSignerID:      *incomingSignerID,
			IncomingKeyID:         *incomingKeyID,
			IncomingActorID:       *incomingActorID,
			IncomingProvisionedAt: *incomingProvisionedAt,
			IncomingRotateBy:      *incomingRotateBy,
			EffectiveAt:           *effectiveAt,
			ReceiptID:             *receiptID,
		}
		if strings.TrimSpace(*incomingRoles) != "" {
			for _, role := range strings.Split(*incomingRoles, ",") {
				role = strings.TrimSpace(role)
				if role != "" {
					request.IncomingRoles = append(request.IncomingRoles, role)
				}
			}
		}
		receipt, err := project.SignerRotationReceipt(bundle, request)
		if err != nil {
			return err
		}
		if *out == "" {
			return printJSON(receipt)
		}
		return project.WriteJSON(filepath.Clean(*out), receipt)
	case "signer-rotation-approve":
		fs := flag.NewFlagSet("signer-rotation-approve", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		receiptPath := fs.String("receipt", "", "signer rotation receipt path")
		role := fs.String("role", "", "approval role")
		signerID := fs.String("signer-id", "", "approval signer id")
		approvedAt := fs.String("approved-at", "", "approval timestamp")
		signatureID := fs.String("signature-id", "", "explicit signature id")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *receiptPath == "" {
			return errors.New("signer-rotation-approve requires --receipt")
		}
		if *role == "" {
			return errors.New("signer-rotation-approve requires --role")
		}
		if *signerID == "" {
			return errors.New("signer-rotation-approve requires --signer-id")
		}
		if *approvedAt == "" {
			return errors.New("signer-rotation-approve requires --approved-at")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		var receipt project.SignerRotationReceiptOutput
		if err := readJSONFile(filepath.Clean(*receiptPath), &receipt); err != nil {
			return err
		}
		approval, err := project.GenerateSignerRotationApproval(bundle, project.SignerRotationApprovalRequest{
			Receipt:      receipt,
			ApprovalRole: *role,
			SignerID:     *signerID,
			ApprovedAt:   *approvedAt,
			SignatureID:  *signatureID,
		})
		if err != nil {
			return err
		}
		if *out == "" {
			return printJSON(approval)
		}
		return project.WriteJSON(filepath.Clean(*out), approval)
	case "signer-rotation-finalize":
		fs := flag.NewFlagSet("signer-rotation-finalize", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		receiptPath := fs.String("receipt", "", "signer rotation receipt path")
		approvalsPath := fs.String("approvals", "", "approval artifact path")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *receiptPath == "" {
			return errors.New("signer-rotation-finalize requires --receipt")
		}
		if *approvalsPath == "" {
			return errors.New("signer-rotation-finalize requires --approvals")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		var receipt project.SignerRotationReceiptOutput
		if err := readJSONFile(filepath.Clean(*receiptPath), &receipt); err != nil {
			return err
		}
		approvals, err := readSignerRotationApprovals(filepath.Clean(*approvalsPath))
		if err != nil {
			return err
		}
		finalized, err := project.SignerRotationFinalize(bundle, project.SignerRotationFinalizeRequest{
			Receipt:   receipt,
			Approvals: approvals,
		})
		if err != nil {
			return err
		}
		if *out == "" {
			return printJSON(finalized)
		}
		return project.WriteJSON(filepath.Clean(*out), finalized)
	case "signer-rotation-activate":
		fs := flag.NewFlagSet("signer-rotation-activate", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		bundlePath := fs.String("bundle", "", "finalized rotation bundle path")
		incomingSharedSecret := fs.String("incoming-shared-secret", "", "incoming signer shared secret")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *bundlePath == "" {
			return errors.New("signer-rotation-activate requires --bundle")
		}
		if *incomingSharedSecret == "" {
			return errors.New("signer-rotation-activate requires --incoming-shared-secret")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		var finalized project.SignerRotationFinalizedBundle
		if err := readJSONFile(filepath.Clean(*bundlePath), &finalized); err != nil {
			return err
		}
		activation, err := project.SignerRotationActivation(bundle, project.SignerRotationActivationRequest{
			FinalizedBundle:      finalized,
			IncomingSharedSecret: *incomingSharedSecret,
		})
		if err != nil {
			return err
		}
		if *out == "" {
			return printJSON(activation)
		}
		return project.WriteJSON(filepath.Clean(*out), activation)
	case "signer-rotation-apply":
		fs := flag.NewFlagSet("signer-rotation-apply", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		planPath := fs.String("plan", "", "activation plan path")
		policyOut := fs.String("policy-out", "", "applied policy output path")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *planPath == "" {
			return errors.New("signer-rotation-apply requires --plan")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		var plan project.SignerRotationActivationPlan
		if err := readJSONFile(filepath.Clean(*planPath), &plan); err != nil {
			return err
		}
		applyResult, err := project.SignerRotationApply(bundle, project.SignerRotationApplyRequest{
			ActivationPlan: plan,
		})
		if err != nil {
			return err
		}
		if *policyOut != "" {
			if err := project.WriteJSON(filepath.Clean(*policyOut), applyResult.AppliedPolicy); err != nil {
				return err
			}
		}
		if *out == "" {
			return printJSON(applyResult)
		}
		return project.WriteJSON(filepath.Clean(*out), applyResult)
	case "signer-rotation-verify":
		fs := flag.NewFlagSet("signer-rotation-verify", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		planPath := fs.String("plan", "", "activation plan path")
		policyPath := fs.String("policy", "", "applied checkpoint signer policy path")
		verifiedAt := fs.String("verified-at", "", "verification timestamp")
		signatureID := fs.String("signature-id", "", "explicit signature id")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *planPath == "" {
			return errors.New("signer-rotation-verify requires --plan")
		}
		if *verifiedAt == "" {
			return errors.New("signer-rotation-verify requires --verified-at")
		}
		bundle, err := project.LoadBundle(*root)
		if err != nil {
			return err
		}
		var plan project.SignerRotationActivationPlan
		if err := readJSONFile(filepath.Clean(*planPath), &plan); err != nil {
			return err
		}
		resolvedPolicyPath := strings.TrimSpace(*policyPath)
		if resolvedPolicyPath == "" {
			resolvedPolicyPath = filepath.Join(*root, filepath.FromSlash(plan.PolicyPath))
		}
		var policy project.CheckpointSignerPolicyOutput
		if err := readJSONFile(filepath.Clean(resolvedPolicyPath), &policy); err != nil {
			return err
		}
		receipt, err := project.SignerRotationVerify(bundle, project.SignerRotationVerifyRequest{
			ActivationPlan: plan,
			Policy:         policy,
			VerifiedAt:     *verifiedAt,
			SignatureID:    *signatureID,
		})
		if err != nil {
			return err
		}
		if *out == "" {
			return printJSON(receipt)
		}
		return project.WriteJSON(filepath.Clean(*out), receipt)
	case "signer-rotation-ledger-append":
		fs := flag.NewFlagSet("signer-rotation-ledger-append", flag.ContinueOnError)
		applyPath := fs.String("apply", "", "signer rotation apply result path")
		verificationPath := fs.String("verification", "", "signer rotation verification receipt path")
		ledgerPath := fs.String("ledger", "", "existing activation audit ledger path")
		ledgerOut := fs.String("ledger-out", "", "updated activation audit ledger output path")
		out := fs.String("out", "", "append result output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *applyPath == "" {
			return errors.New("signer-rotation-ledger-append requires --apply")
		}
		if *verificationPath == "" {
			return errors.New("signer-rotation-ledger-append requires --verification")
		}
		var applyResult project.SignerRotationApplyResult
		if err := readJSONFile(filepath.Clean(*applyPath), &applyResult); err != nil {
			return err
		}
		var verification project.SignerRotationVerificationReceipt
		if err := readJSONFile(filepath.Clean(*verificationPath), &verification); err != nil {
			return err
		}
		ledger := project.SignerRotationActivationAuditLedger{}
		if strings.TrimSpace(*ledgerPath) != "" {
			contents, err := os.ReadFile(filepath.Clean(*ledgerPath))
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return err
				}
			} else if len(contents) > 0 {
				if err := json.Unmarshal(contents, &ledger); err != nil {
					return fmt.Errorf("parse %s: %w", *ledgerPath, err)
				}
			}
		}
		appendResult, err := project.SignerRotationActivationAuditAppend(project.SignerRotationActivationAuditAppendRequest{
			ApplyResult:         applyResult,
			VerificationReceipt: verification,
			ExistingLedger:      ledger,
		})
		if err != nil {
			return err
		}
		if *ledgerOut != "" {
			if err := project.WriteJSON(filepath.Clean(*ledgerOut), appendResult.Ledger); err != nil {
				return err
			}
		}
		if *out == "" {
			return printJSON(appendResult)
		}
		return project.WriteJSON(filepath.Clean(*out), appendResult)
	case "signer-rotation-ledger-reconcile":
		fs := flag.NewFlagSet("signer-rotation-ledger-reconcile", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		ledgerPath := fs.String("ledger", "", "activation audit ledger path")
		policyPath := fs.String("policy", "", "checkpoint signer policy path")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		resolvedLedgerPath := strings.TrimSpace(*ledgerPath)
		if resolvedLedgerPath == "" {
			resolvedLedgerPath = filepath.Join(*root, "build/rotation/activation-audit-ledger.json")
		}
		resolvedPolicyPath := strings.TrimSpace(*policyPath)
		if resolvedPolicyPath == "" {
			resolvedPolicyPath = filepath.Join(*root, "config/governance/checkpoint-signers.json")
		}
		ledger, err := readActivationAuditLedger(resolvedLedgerPath)
		if err != nil {
			return err
		}
		var policy project.CheckpointSignerPolicyOutput
		if err := readJSONFile(filepath.Clean(resolvedPolicyPath), &policy); err != nil {
			return err
		}
		report, err := project.SignerRotationActivationAuditReconcile(project.SignerRotationActivationAuditReconcileRequest{
			Ledger:     ledger,
			Policy:     policy,
			PolicyPath: checkpointSignerPolicyPath,
		})
		if err != nil {
			return err
		}
		if *out == "" {
			return printJSON(report)
		}
		return project.WriteJSON(filepath.Clean(*out), report)
	case "signer-rotation-ledger-export":
		fs := flag.NewFlagSet("signer-rotation-ledger-export", flag.ContinueOnError)
		root := fs.String("root", ".", "project root")
		ledgerPath := fs.String("ledger", "", "activation audit ledger path")
		policyPath := fs.String("policy", "", "checkpoint signer policy path")
		reconcilePath := fs.String("reconcile", "", "existing reconciliation report path")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		resolvedLedgerPath := strings.TrimSpace(*ledgerPath)
		if resolvedLedgerPath == "" {
			resolvedLedgerPath = filepath.Join(*root, "build/rotation/activation-audit-ledger.json")
		}
		resolvedPolicyPath := strings.TrimSpace(*policyPath)
		if resolvedPolicyPath == "" {
			resolvedPolicyPath = filepath.Join(*root, "config/governance/checkpoint-signers.json")
		}
		ledger, err := readActivationAuditLedger(resolvedLedgerPath)
		if err != nil {
			return err
		}
		var policy project.CheckpointSignerPolicyOutput
		if err := readJSONFile(filepath.Clean(resolvedPolicyPath), &policy); err != nil {
			return err
		}
		var report project.SignerRotationActivationAuditReconcileReport
		if strings.TrimSpace(*reconcilePath) != "" {
			if err := readJSONFile(filepath.Clean(*reconcilePath), &report); err != nil {
				return err
			}
		} else {
			report, err = project.SignerRotationActivationAuditReconcile(project.SignerRotationActivationAuditReconcileRequest{
				Ledger:     ledger,
				Policy:     policy,
				PolicyPath: checkpointSignerPolicyPath,
			})
			if err != nil {
				return err
			}
		}
		exportPackage, err := project.SignerRotationActivationAuditExport(project.SignerRotationActivationAuditExportRequest{
			Ledger:         ledger,
			Policy:         policy,
			Reconciliation: report,
			PolicyPath:     checkpointSignerPolicyPath,
		})
		if err != nil {
			return err
		}
		if *out == "" {
			return printJSON(exportPackage)
		}
		return project.WriteJSON(filepath.Clean(*out), exportPackage)
	case "signer-rotation-ledger-verify-export":
		fs := flag.NewFlagSet("signer-rotation-ledger-verify-export", flag.ContinueOnError)
		exportPath := fs.String("export", "", "activation audit export package path")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *exportPath == "" {
			return errors.New("signer-rotation-ledger-verify-export requires --export")
		}
		var exportPackage project.SignerRotationActivationAuditExportPackage
		if err := readJSONFile(filepath.Clean(*exportPath), &exportPackage); err != nil {
			return err
		}
		report, err := project.SignerRotationActivationAuditVerifyExport(project.SignerRotationActivationAuditExportVerificationRequest{
			ExportPackage: exportPackage,
		})
		if err != nil {
			return err
		}
		if *out != "" {
			if err := project.WriteJSON(filepath.Clean(*out), report); err != nil {
				return err
			}
		} else if err := printJSON(report); err != nil {
			return err
		}
		if report.Status == "invalid" {
			return errors.New("activation audit export verification failed")
		}
		return nil
	case "signer-rotation-ledger-archive-index":
		fs := flag.NewFlagSet("signer-rotation-ledger-archive-index", flag.ContinueOnError)
		exportPaths := fs.String("exports", "", "comma-separated activation audit export package paths")
		out := fs.String("out", "", "output file path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*exportPaths) == "" {
			return errors.New("signer-rotation-ledger-archive-index requires --exports")
		}
		packages, err := readActivationAuditExportPackages(*exportPaths)
		if err != nil {
			return err
		}
		index, err := project.BuildSignerRotationActivationAuditArchiveIndex(project.SignerRotationActivationAuditArchiveIndexRequest{
			Packages: packages,
		})
		if err != nil {
			return err
		}
		if *out != "" {
			if err := project.WriteJSON(filepath.Clean(*out), index); err != nil {
				return err
			}
		} else if err := printJSON(index); err != nil {
			return err
		}
		if index.Status == "invalid" {
			return errors.New("activation audit archive index verification failed")
		}
		return nil
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
		"usage: 0aid <version|module-map|module-plan|identity-plan|signer-manifest|signer-rotation-receipt|signer-rotation-approve|signer-rotation-finalize|signer-rotation-activate|signer-rotation-apply|signer-rotation-verify|signer-rotation-ledger-append|signer-rotation-ledger-reconcile|signer-rotation-ledger-export|signer-rotation-ledger-verify-export|signer-rotation-ledger-archive-index|show-plan|init-genesis|render-validator|render-identity|init-node|collect-validator|assemble-genesis|assemble-localnet> [flags]",
	)
}

func readJSONFile(path string, destination any) error {
	contents, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(contents, destination); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

func readActivationAuditLedger(path string) (project.SignerRotationActivationAuditLedger, error) {
	ledger := project.SignerRotationActivationAuditLedger{}
	contents, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ledger, nil
		}
		return ledger, err
	}
	if len(contents) == 0 {
		return ledger, nil
	}
	if err := json.Unmarshal(contents, &ledger); err != nil {
		return ledger, fmt.Errorf("parse %s: %w", path, err)
	}
	return ledger, nil
}

func readActivationAuditExportPackages(path string) ([]project.SignerRotationActivationAuditArchivePackage, error) {
	paths := strings.Split(path, ",")
	collected := make([]project.SignerRotationActivationAuditArchivePackage, 0, len(paths))
	for _, rawPath := range paths {
		cleanPath := filepath.Clean(strings.TrimSpace(rawPath))
		if cleanPath == "" {
			continue
		}
		var exportPackage project.SignerRotationActivationAuditExportPackage
		if err := readJSONFile(cleanPath, &exportPackage); err != nil {
			return nil, err
		}
		collected = append(collected, project.SignerRotationActivationAuditArchivePackage{
			PackagePath:   filepath.ToSlash(cleanPath),
			ExportPackage: exportPackage,
		})
	}
	if len(collected) == 0 {
		return nil, fmt.Errorf("no activation audit export packages found in %s", path)
	}
	return collected, nil
}

func readSignerRotationApprovals(path string) ([]project.SignerRotationApproval, error) {
	paths := strings.Split(path, ",")
	collected := make([]project.SignerRotationApproval, 0, len(paths))
	for _, rawPath := range paths {
		cleanPath := filepath.Clean(strings.TrimSpace(rawPath))
		if cleanPath == "" {
			continue
		}
		contents, err := os.ReadFile(cleanPath)
		if err != nil {
			return nil, err
		}
		var envelope project.SignerRotationApprovalEnvelope
		if err := json.Unmarshal(contents, &envelope); err == nil && len(envelope.Approvals) > 0 {
			collected = append(collected, envelope.Approvals...)
			continue
		}
		var approvals []project.SignerRotationApproval
		if err := json.Unmarshal(contents, &approvals); err == nil && len(approvals) > 0 {
			collected = append(collected, approvals...)
			continue
		}
		var single project.SignerRotationApproval
		if err := json.Unmarshal(contents, &single); err == nil && single.ReceiptID != "" {
			collected = append(collected, single)
			continue
		}
		return nil, fmt.Errorf("parse %s: expected approval object, array, or envelope", cleanPath)
	}
	if len(collected) == 0 {
		return nil, fmt.Errorf("no approval artifacts found in %s", path)
	}
	return collected, nil
}

func printJSON(value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", encoded)
	return nil
}
