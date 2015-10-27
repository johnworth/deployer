package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var (
	externalRepo   = flag.String("git-repo-external", "", "The external git repository to clone")
	externalBranch = flag.String("git-branch-external", "", "The git branch to check out in the external repo")
	gitRepo        = flag.String("git-repo-internal", "", "The internal git repository to clone")
	gitBranch      = flag.String("git-branch-internal", "dev", "The git branch to check out in the internal repo")
	account        = flag.String("account", "discoenv", "The Docker account to use")
	repo           = flag.String("repo", "", "The Docker repo to pull")
	vaultPass      = flag.String("vault-pass", "", "The path to the ansible vault password file")
	secretFile     = flag.String("secret", "", "The file encrypted by ansible-vault")
	inventory      = flag.String("inventory", "", "The ansible inventory to use")
	tag            = flag.String("tag", "dev", "The docker tag to pull from")
	user           = flag.String("user", "", "The sudo user to use with the ansible command")
	service        = flag.String("service", "", "The service to restart on the host")
	configTag      = flag.String("config-tag", "", "The ansible tag to pass to the ansible-playbook command when updating the configs")
	pullTag        = flag.String("pull-tag", "", "The ansible tag to pass to the ansible-playbook command when updating the images")
	serviceTag     = flag.String("service-tag", "", "The ansible tag to pass to the ansible-playbook command when updating systemd service files")
	restartTag     = flag.String("restart-tag", "", "The ansible tag to pass to the ansible-playbook command when restarting the containers")
	playbook       = flag.String("playbook", "", "The ansible playbook to use")
)

const (
	internalDir = "internal-deployer-checkout"
	externalDir = "external-deployer-checkout"
)

func init() {
	flag.Parse()
}

func main() {
	if *gitRepo == "" {
		fmt.Println("--git-repo-internal must be set.")
		os.Exit(-1)
	}

	if *externalRepo == "" {
		fmt.Println("--git-repo-external must be set.")
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

	if *pullTag == "" {
		fmt.Println("--pull-tag must be set")
		os.Exit(-1)
	}

	if *restartTag == "" {
		fmt.Println("--restart-tag must be set")
		os.Exit(-1)
	}

	if *serviceTag == "" {
		fmt.Println("--service-tag must be set")
		os.Exit(-1)
	}

	if *playbook == "" {
		fmt.Println("--playbook must be set")
		os.Exit(-1)
	}

	git, err := exec.LookPath("git")
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	ansiblePlaybook, err := exec.LookPath("ansible-playbook")
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	if _, err := os.Stat(internalDir); err == nil {
		if err = os.RemoveAll(internalDir); err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}
	}

	if _, err := os.Stat(externalDir); err == nil {
		if err = os.RemoveAll(externalDir); err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}
	}

	fmt.Printf("Cloning the internal repo %s \n", *gitRepo)
	cmd := exec.Command(git, "clone", *gitRepo, internalDir)
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("Cloning the external repo %s\n", *externalRepo)
	cmd = exec.Command(git, "clone", *externalRepo, externalDir)
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	origDir, err := os.Getwd()
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("cd'ing into %s\n", internalDir)
	err = os.Chdir(internalDir)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("Checking out the %s branch from the internal repo\n", *gitBranch)
	cmd = exec.Command(git, "checkout", *gitBranch)
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("Pulling the %s branch from the internal repo\n", *gitBranch)
	cmd = exec.Command(git, "pull")
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("cd'ing into %s\n", origDir)
	err = os.Chdir(origDir)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("cd'ing into %s\n", externalDir)
	err = os.Chdir(externalDir)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("Checking out the %s branch from the external repo\n", *externalBranch)
	cmd = exec.Command(git, "checkout", *externalBranch)
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("Pulling the %s branch from the external repo\n", *externalBranch)
	cmd = exec.Command(git, "pull")
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("cd'ing into %s\n", origDir)
	err = os.Chdir(origDir)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	internalGroupVars, err := filepath.Abs(path.Join(internalDir, "group_vars"))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}
	externalGroupVars, err := filepath.Abs(path.Join(externalDir, "group_vars"))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("Copying files from %s to %s\n", internalGroupVars, externalGroupVars)
	var copyPaths []string

	visit := func(p string, i os.FileInfo, err error) error {
		if !i.IsDir() {
			fmt.Printf("Found file %s to copy\n", p)
			copyPaths = append(copyPaths, p)
		}
		return err
	}

	err = filepath.Walk(internalGroupVars, visit)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	if _, err := os.Stat(externalGroupVars); os.IsNotExist(err) {
		fmt.Printf("Creating %s\n", externalGroupVars)
		err = os.MkdirAll(externalGroupVars, 0755)
		if err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}
	}

	for _, copyPath := range copyPaths {
		destPath := path.Join(externalGroupVars, path.Base(copyPath))
		fmt.Printf("Copying %s to %s\n", copyPath, destPath)
		contents, err := ioutil.ReadFile(copyPath)
		if err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}
		err = ioutil.WriteFile(destPath, contents, 0755)
		if err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}
	}

	internalInventories, err := filepath.Abs(path.Join(internalDir, "inventories"))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}
	externalInventories, err := filepath.Abs(path.Join(externalDir, "inventories"))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("Copying files from %s to %s\n", internalInventories, externalInventories)
	copyPaths = []string{}

	err = filepath.Walk(internalInventories, visit)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	for _, copyPath := range copyPaths {
		destPath := path.Join(externalInventories, path.Base(copyPath))
		fmt.Printf("Copying %s to %s\n", copyPath, destPath)
		contents, err := ioutil.ReadFile(copyPath)
		if err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}
		err = ioutil.WriteFile(destPath, contents, 0755)
		if err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}
	}

	fmt.Printf("cd'ing into %s\n", externalDir)
	err = os.Chdir(externalDir)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("Updating %s/%s:%s with ansible\n", *account, *repo, *tag)
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
		*pullTag,
		*playbook,
	)
	fmt.Printf("%s %s\n", cmd.Path, strings.Join(cmd.Args, " "))
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Print(err)
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
		*playbook,
	)
	fmt.Printf("%s %s\n", cmd.Path, strings.Join(cmd.Args, " "))
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("Updating service file for %s with ansible\n", *repo)
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
		*serviceTag,
		*playbook,
	)
	fmt.Printf("%s %s\n", cmd.Path, strings.Join(cmd.Args, " "))
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	fmt.Printf("Restarting %s with ansible\n", *repo)
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
		*restartTag,
		*playbook,
	)
	fmt.Printf("%s %s\n", cmd.Path, strings.Join(cmd.Args, " "))
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output[:]))
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}
}
