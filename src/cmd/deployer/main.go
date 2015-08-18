package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	gitRepo        = flag.String("git-repo", "", "The git repository to clone")
	gitBranch      = flag.String("git-branch", "dev", "The git branch to checkout")
	account        = flag.String("account", "discoenv", "The Docker account to use")
	repo           = flag.String("repo", "", "The Docker repo to pull")
	vaultPass      = flag.String("vault-pass", "", "The path to the ansible vault password file")
	secretFile     = flag.String("secret", "", "The file encrypted by ansible-vault")
	inventory      = flag.String("inventory", "", "The ansible inventory to use")
	tag            = flag.String("tag", "dev", "The docker tag to pull from")
	user           = flag.String("user", "", "The sudo user to use with the ansible command")
	service        = flag.String("service", "", "The service to restart on the host")
	configTag      = flag.String("config-tag", "", "The ansible tag to pass to the ansible-playbook command when updating the configs")
	configPlaybook = flag.String("config-playbook", "", "The playbook to deploy configs through")
)

func init() {
	flag.Parse()
}

func main() {
	if *gitRepo == "" {
		fmt.Println("--git-repo must be set.")
		os.Exit(-1)
	}

	if *account == "" {
		fmt.Println("--account must be set.")
		os.Exit(-1)
	}

	if *repo == "" {
		fmt.Println("--repo must be set.")
		os.Exit(-1)
	}

	if *vaultPass == "" {
		fmt.Println("--vault-pass must be set.")
		os.Exit(-1)
	}

	if *secretFile == "" {
		fmt.Println("--secret must be set")
		os.Exit(-1)
	}

	if *inventory == "" {
		fmt.Println("--inventory must be set")
		os.Exit(-1)
	}

	if *user == "" {
		fmt.Println("--user must be set")
		os.Exit(-1)
	}

	if *service == "" {
		fmt.Println("--service must be set")
		os.Exit(-1)
	}

	if *configTag == "" {
		fmt.Println("--config-tag must be set")
		os.Exit(-1)
	}

	if *configPlaybook == "" {
		fmt.Println("--config-playbook must be set")
		os.Exit(-1)
	}

	git, err := exec.LookPath("git")
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}

	ansible, err := exec.LookPath("ansible")
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}

	ansiblePlaybook, err := exec.LookPath("ansible-playbook")
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}

	if _, err := os.Stat("deployer-checkout"); err == nil {
		if err = os.RemoveAll("deployer-checkout"); err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(-1)
		}
	}

	fmt.Printf("Cloning %s \n", *gitRepo)
	cmd := exec.Command(git, "clone", *gitRepo, "deployer-checkout")
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}

	fmt.Println("cd'ing into deployer-checkout")
	err = os.Chdir("deployer-checkout")
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}

	fmt.Printf("Checking out the %s branch\n", *gitBranch)
	cmd = exec.Command(git, "checkout", *gitBranch)
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}

	fmt.Printf("Pulling the %s branch\n", *gitBranch)
	cmd = exec.Command(git, "pull")
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}

	fmt.Printf("Updating %s/%s:%s with ansible\n", *account, *repo, *tag)
	cmd = exec.Command(
		ansible,
		"services",
		"-e",
		fmt.Sprintf("@%s", *secretFile),
		fmt.Sprintf("--vault-password-file=%s", *vaultPass),
		"-i",
		*inventory,
		"--sudo",
		"-u",
		*user,
		"-a",
		fmt.Sprintf("docker pull %s/%s:%s", *account, *repo, *tag),
	)
	fmt.Printf("%s %s\n", cmd.Path, strings.Join(cmd.Args, " "))
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}

	fmt.Printf("Configuring %s with ansible\n", *repo)
	cmd = exec.Command(
		ansiblePlaybook,
		"-e",
		fmt.Sprintf("@%s", *secretFile),
		fmt.Sprintf("--vault-password-file=%s", *vaultPass),
		"-i",
		*inventory,
		"--sudo",
		"-u",
		*user,
		"--tags",
		*configTag,
		*configPlaybook,
	)
	fmt.Printf("%s %s\n", cmd.Path, strings.Join(cmd.Args, " "))
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}

	fmt.Printf("Restarting %s\n", *repo)
	cmd = exec.Command(
		ansible,
		"services",
		"-e",
		fmt.Sprintf("@%s", *secretFile),
		fmt.Sprintf("--vault-password-file=%s", *vaultPass),
		"-i",
		*inventory,
		"--sudo",
		"-u",
		*user,
		"-a",
		fmt.Sprintf("systemctl restart %s", *service),
	)
	fmt.Printf("%s %s\n", cmd.Path, strings.Join(cmd.Args, " "))
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(-1)
	}
}
