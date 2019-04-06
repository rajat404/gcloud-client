package action

import (
	"context"
	"fmt"

	"github.com/devdinu/gcloud-client/command"
	"github.com/devdinu/gcloud-client/config"
	"github.com/devdinu/gcloud-client/gcloud"
	"github.com/devdinu/gcloud-client/logger"
)

type InstanceLogin struct {
	ctx context.Context
	f   finder
}

func (il InstanceLogin) Login(c gcloud.Client, args config.Args) error {
	projs, err := listProjects(il.ctx, c)
	if err != nil {
		return err
	}
	pattern := args.InstanceCmdArgs.Prefix
	//TODO: override with commandline projects to reduce search space
	insts, err := il.f.Search(il.ctx, projs.Names(), pattern)
	if err != nil {
		fmt.Printf("[Search] couldn't search instances with prefix %s err: %v", pattern, err)
		return err
	}
	fmt.Println("Search By Prefix Result: ")
	for _, ins := range insts {
		logger.Infof("%s: name: %s\tip: %s\t", ins.Project, ins.Name, ins.IP())
	}
	tmuxCfg := command.TmuxConfig{
		Project: "ssh_instances_pane_cmd",
		Session: args.Login.Session,
	}
	tmuxCfg.AddArg("user", args.Login.User)
	_, err = c.Login(il.ctx, insts, "hostname", tmuxCfg)
	return err
}

func NewLogin(ctx context.Context, f finder) InstanceLogin {
	return InstanceLogin{ctx: ctx, f: f}
}