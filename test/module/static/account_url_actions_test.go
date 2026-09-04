package static_test

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/test/harness"
)

func TestAccountReportActionsAreURLActions(t *testing.T) {
	root := harness.RepoRoot(t)
	reportXML := filepath.Join(filepath.Dir(root), "sumeru_addons", "account", "views", "report_actions.xml")
	if _, err := os.Stat(reportXML); err != nil {
		t.Skipf("account report_actions.xml not present at %s", reportXML)
	}
	raw, err := os.ReadFile(reportXML)
	if err != nil {
		t.Fatal(err)
	}
	var vl parser.ViewList
	if err := xml.Unmarshal(raw, &vl); err != nil {
		t.Fatal(err)
	}
	vl.MergeViewListData()
	want := map[string]string{
		"action_report_profit_loss":      "/account/reports/view?type=profit_loss",
		"action_report_balance_sheet":    "/account/reports/view?type=balance_sheet",
		"action_report_trial_balance":    "/account/reports/view?type=trial_balance",
		"action_report_general_ledger":   "/account/reports/view?type=general_ledger",
		"action_report_partner_ledger":   "/account/reports/view?type=partner_ledger",
		"action_report_aged_receivable":  "/account/reports/view?type=aged_receivable",
		"action_report_aged_payable":     "/account/reports/view?type=aged_payable",
		"action_report_cash_flow":        "/account/reports/view?type=cash_flow",
		"action_report_annual_composite": "/account/reports/view?type=annual_composite",
	}
	got := map[string]string{}
	for _, rec := range vl.Records {
		if rec.Model != "sys.action.url" {
			t.Fatalf("record %s model=%q want sys.action.url", rec.ID, rec.Model)
		}
		var url string
		for _, f := range rec.Field {
			if f.Name == "url" {
				url = f.Body
			}
		}
		got[rec.ID] = url
	}
	for id, path := range want {
		if got[id] != path {
			t.Fatalf("%s url=%q want %q", id, got[id], path)
		}
	}
}

func TestAccountBankReconcileIsURLAction(t *testing.T) {
	root := harness.RepoRoot(t)
	bankXML := filepath.Join(filepath.Dir(root), "sumeru_addons", "account", "views", "account_bank_actions.xml")
	if _, err := os.Stat(bankXML); err != nil {
		t.Skipf("account bank actions not present at %s", bankXML)
	}
	raw, err := os.ReadFile(bankXML)
	if err != nil {
		t.Fatal(err)
	}
	var vl parser.ViewList
	if err := xml.Unmarshal(raw, &vl); err != nil {
		t.Fatal(err)
	}
	vl.MergeViewListData()
	var found bool
	for _, rec := range vl.Records {
		if rec.ID != "action_bank_reconcile_workspace" {
			continue
		}
		found = true
		if rec.Model != "sys.action.url" {
			t.Fatalf("model=%q want sys.action.url", rec.Model)
		}
		var url string
		for _, f := range rec.Field {
			if f.Name == "url" {
				url = f.Body
			}
		}
		if url != "/account/bank/reconcile" {
			t.Fatalf("url=%q", url)
		}
	}
	if !found {
		t.Fatal("action_bank_reconcile_workspace not found")
	}
}
