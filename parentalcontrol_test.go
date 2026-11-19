package glinet_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

// sentRaw records "method params" of each call of the methods of module on srv, in call order.
func sentRaw(srv *glinettest.Server, module string, methods ...string) *[]string {
	var sent []string
	for _, method := range methods {
		next := srv.Handler(module, method)
		srv.Handle(module, method, func(params json.RawMessage) (any, error) {
			sent = append(sent, method+" "+string(params))
			return next(params)
		})
	}
	return &sent
}

func checkSent(t *testing.T, sent *[]string, want ...string) {
	t.Helper()
	if !slices.Equal(*sent, want) {
		t.Errorf("params =\n%q\nwant\n%q", *sent, want)
	}
	*sent = nil
}

func TestParentalControlConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentRaw(srv, "parental-control", "set_config")

	got, err := c.ParentalControl.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if got.Enable || len(got.Groups) != 0 || len(got.Rules) != 2 || got.Rules[0].ID != glinet.ParentalControlRuleDrop || !got.Rules[1].Preset {
		t.Errorf("GetConfig = %+v", *got)
	}

	if err := c.ParentalControl.SetConfig(ctx, glinet.ParentalControlSettings{Enable: true, DropAnonymous: true}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	checkSent(t, sent, `set_config {"enable":true,"drop_anonymous":true,"auto_update":false}`)

	got, err = c.ParentalControl.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !got.Enable || !got.DropAnonymous || got.AutoUpdate {
		t.Errorf("GetConfig = %+v", *got)
	}
}

func TestParentalControlRules(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentRaw(srv, "parental-control", "add_rule", "set_rule", "remove_rule")

	manual, err := c.ParentalControl.AddRule(ctx, glinet.ParentalControlRuleSettings{
		Name: "example", Color: "#417505", Apps: []int{1001, 2}, Manual: true, Blacklist: []string{"example.com", ""},
	})
	if err != nil {
		t.Fatalf("AddRule: %v", err)
	}
	url, err := c.ParentalControl.AddRule(ctx, glinet.ParentalControlRuleSettings{Name: "list", Color: "#d0021b", URL: "https://example.com/list.txt"})
	if err != nil {
		t.Fatalf("AddRule url: %v", err)
	}
	if manual == "" || url == "" || manual == url {
		t.Fatalf("AddRule IDs = %q, %q", manual, url)
	}
	if _, err := c.ParentalControl.AddRule(ctx, glinet.ParentalControlRuleSettings{Name: "empty", Manual: true}); err != nil {
		t.Fatalf("AddRule empty: %v", err)
	}
	checkSent(t, sent,
		`add_rule {"name":"example","color":"#417505","apps":[1001,2],"manual":true,"blacklist":["example.com",""]}`,
		`add_rule {"name":"list","color":"#d0021b","apps":[],"manual":false,"url":"https://example.com/list.txt"}`,
		`add_rule {"name":"empty","color":"","apps":[],"manual":true,"blacklist":[]}`,
	)

	if err := c.ParentalControl.SetRule(ctx, manual, glinet.ParentalControlRuleSettings{Name: "example-2", Color: "#417505", Manual: true, Blacklist: []string{"example.org"}}); err != nil {
		t.Fatalf("SetRule: %v", err)
	}
	checkSent(t, sent, `set_rule {"id":"`+manual+`","name":"example-2","color":"#417505","apps":[],"manual":true,"blacklist":["example.org"]}`)

	cfg, err := c.ParentalControl.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	i := slices.IndexFunc(cfg.Rules, func(r glinet.ParentalControlRule) bool { return r.ID == manual })
	if i < 0 || cfg.Rules[i].Name != "example-2" || cfg.Rules[i].Preset || !slices.Equal(cfg.Rules[i].Blacklist, []string{"example.org"}) {
		t.Errorf("GetConfig rules = %+v", cfg.Rules)
	}

	if err := c.ParentalControl.RemoveRule(ctx, url); err != nil {
		t.Fatalf("RemoveRule: %v", err)
	}
	checkSent(t, sent, `remove_rule {"id":"`+url+`"}`)
}

func TestParentalControlGroups(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentRaw(srv, "parental-control", "add_group", "set_group", "remove_group", "set_brief", "get_brief")

	id, err := c.ParentalControl.AddGroup(ctx, glinet.ParentalControlGroupSettings{Name: "example", DefaultRule: glinet.ParentalControlRuleAccept, MACs: []string{"02:00:00:00:00:01"}})
	if err != nil {
		t.Fatalf("AddGroup: %v", err)
	}
	empty, err := c.ParentalControl.AddGroup(ctx, glinet.ParentalControlGroupSettings{Name: "empty", DefaultRule: glinet.ParentalControlRuleDrop})
	if err != nil {
		t.Fatalf("AddGroup empty: %v", err)
	}
	checkSent(t, sent,
		`add_group {"name":"example","default_rule":"accept","macs":["02:00:00:00:00:01"]}`,
		`add_group {"name":"empty","default_rule":"drop","macs":[]}`,
	)

	schedules := []glinet.ParentalControlSchedule{
		{Week: 1, Begin: "08:00", End: "12:00", Rule: glinet.ParentalControlRuleDrop},
		{Week: 6, Begin: "20:00", End: "24:00", Rule: glinet.ParentalControlRuleDrop},
	}
	if err := c.ParentalControl.SetSchedules(ctx, id, schedules); err != nil {
		t.Fatalf("SetSchedules: %v", err)
	}
	if err := c.ParentalControl.SetGroup(ctx, id, glinet.ParentalControlGroupSettings{Name: "example-2", DefaultRule: glinet.ParentalControlRuleDrop}); err != nil {
		t.Fatalf("SetGroup: %v", err)
	}
	checkSent(t, sent,
		`set_group {"id":"`+id+`","schedules":[{"week":1,"begin":"08:00","end":"12:00","rule":"drop"},{"week":6,"begin":"20:00","end":"24:00","rule":"drop"}]}`,
		`set_group {"id":"`+id+`","name":"example-2","default_rule":"drop","macs":[]}`,
	)

	cfg, err := c.ParentalControl.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	g := cfg.Groups[0]
	if g.ID != id || g.Name != "example-2" || g.DefaultRule != glinet.ParentalControlRuleDrop || len(g.MACs) != 0 || len(g.Schedules) != 2 || g.Schedules[0].ID == "" {
		t.Fatalf("GetConfig groups = %+v", cfg.Groups)
	}

	kept := g.Schedules[:1]
	if err := c.ParentalControl.SetSchedules(ctx, id, kept); err != nil {
		t.Fatalf("SetSchedules: %v", err)
	}
	if err := c.ParentalControl.SetSchedules(ctx, empty, nil); err != nil {
		t.Fatalf("SetSchedules empty: %v", err)
	}
	checkSent(t, sent,
		`set_group {"id":"`+id+`","schedules":[{"id":"`+kept[0].ID+`","week":1,"begin":"08:00","end":"12:00","rule":"drop"}]}`,
		`set_group {"id":"`+empty+`","schedules":[]}`,
	)

	if err := c.ParentalControl.StartBrief(ctx, id, glinet.ParentalControlRuleAccept, "21:15"); err != nil {
		t.Fatalf("StartBrief: %v", err)
	}
	brief, err := c.ParentalControl.GetBrief(ctx, id)
	if err != nil {
		t.Fatalf("GetBrief: %v", err)
	}
	if !brief.Enable || brief.ManualStop || brief.Time != "21:15" || brief.RuleID != glinet.ParentalControlRuleAccept {
		t.Errorf("GetBrief = %+v", *brief)
	}
	status, err := c.ParentalControl.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	want := []glinet.ParentalControlGroupStatus{{ID: id, Rule: "accept", Brief: true}, {ID: empty, Rule: "drop"}}
	if !status.TimeValid || !slices.Equal(status.Groups, want) {
		t.Errorf("GetStatus = %+v, want groups %+v", *status, want)
	}
	if err := c.ParentalControl.RemoveRule(ctx, glinet.ParentalControlRuleAccept); rpcCode(t, err) != -3 {
		t.Errorf("RemoveRule of the temporary ruleset: %v, want error -3", err)
	}
	if err := c.ParentalControl.RemoveRule(ctx, glinet.ParentalControlRuleDrop); rpcCode(t, err) != -2 {
		t.Errorf("RemoveRule of a default ruleset: %v, want error -2", err)
	}

	if err := c.ParentalControl.StartBrief(ctx, empty, glinet.ParentalControlRuleAccept, ""); err != nil {
		t.Fatalf("StartBrief manual stop: %v", err)
	}
	if err := c.ParentalControl.StopBrief(ctx, id); err != nil {
		t.Fatalf("StopBrief: %v", err)
	}
	brief, err = c.ParentalControl.GetBrief(ctx, id)
	if err != nil {
		t.Fatalf("GetBrief: %v", err)
	}
	if brief.Enable {
		t.Errorf("GetBrief after StopBrief = %+v", *brief)
	}
	checkSent(t, sent,
		`set_brief {"enable":true,"manual_stop":false,"rule_id":"accept","group_id":"`+id+`","time":"21:15"}`,
		`get_brief {"group_id":"`+id+`"}`,
		`set_brief {"enable":true,"manual_stop":true,"rule_id":"accept","group_id":"`+empty+`"}`,
		`set_brief {"enable":false,"group_id":"`+id+`"}`,
		`get_brief {"group_id":"`+id+`"}`,
	)

	if err := c.ParentalControl.RemoveGroup(ctx, id); err != nil {
		t.Fatalf("RemoveGroup: %v", err)
	}
	checkSent(t, sent, `remove_group {"id":"`+id+`"}`)
	if err := c.ParentalControl.RemoveGroup(ctx, id); rpcCode(t, err) != glinet.CodeInvalidParams {
		t.Errorf("RemoveGroup of a removed group: %v", err)
	}
}

func TestParentalControlMode(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentRaw(srv, "parental-control", "set_mode")

	if mode, err := c.ParentalControl.GetMode(ctx); err != nil || mode != glinet.ParentalControlModeDefault {
		t.Fatalf("GetMode = %d, %v", mode, err)
	}
	if err := c.ParentalControl.SetMode(ctx, glinet.ParentalControlModeBark); err != nil {
		t.Fatalf("SetMode: %v", err)
	}
	checkSent(t, sent, `set_mode {"mode":2}`)
	if mode, err := c.ParentalControl.GetMode(ctx); err != nil || mode != glinet.ParentalControlModeBark {
		t.Errorf("GetMode = %d, %v", mode, err)
	}
}

func TestParentalControlListApps(t *testing.T) {
	c, _ := newClient(t)

	apps, err := c.ParentalControl.ListApps(context.Background())
	if err != nil {
		t.Fatalf("ListApps: %v", err)
	}
	if len(apps) != 4 || apps[0] != (glinet.ParentalControlApp{ID: 1001, Name: "QQ", Type: 1}) {
		t.Errorf("ListApps = %+v", apps)
	}
}

func TestParentalControlNewIDsSkipLoadedIDs(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	srv.ParentalControl.SetConfig(json.RawMessage(`{
		"enable": false,
		"groups": [{"id": "cfg000001", "name": "loaded", "default_rule": "cfg000003", "macs": [],
			"schedules": [{"id": "cfg000002", "week": 1, "begin": "08:00", "end": "12:00", "rule": "drop"}]}],
		"rules": [{"id": "cfg000003", "name": "loaded", "preset": false}]
	}`))

	group, err := c.ParentalControl.AddGroup(ctx, glinet.ParentalControlGroupSettings{Name: "new", DefaultRule: glinet.ParentalControlRuleDrop})
	if err != nil {
		t.Fatalf("AddGroup: %v", err)
	}
	rule, err := c.ParentalControl.AddRule(ctx, glinet.ParentalControlRuleSettings{Name: "new"})
	if err != nil {
		t.Fatalf("AddRule: %v", err)
	}
	for _, id := range []string{group, rule} {
		if slices.Contains([]string{"cfg000001", "cfg000002", "cfg000003"}, id) {
			t.Errorf("new ID %q is the ID of a loaded item", id)
		}
	}

	if err := c.ParentalControl.RemoveGroup(ctx, group); err != nil {
		t.Fatalf("RemoveGroup: %v", err)
	}
	cfg, err := c.ParentalControl.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if len(cfg.Groups) != 1 || cfg.Groups[0].ID != "cfg000001" {
		t.Errorf("groups after RemoveGroup of the new group = %+v, want the loaded group", cfg.Groups)
	}
}
