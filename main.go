package main

import (
	"fmt"
	"os"
	"time"

	"github.com/skippbox/libretto/virtualmachine/exoscale"
)

func main() {

	key := os.Getenv("EXOSCALE_API_KEY")
	secret := os.Getenv("EXOSCALE_API_SECRET")

	if key == "" || secret == "" {
		usage()
		os.Exit(2)
	}

	vm := createExoscalVM(key, secret)

	fmt.Println(">>About to provission")
	err := vm.Provision()
	if err != nil {
		fmt.Printf("Error provisioning machine: %s\n", err)
	}

	fmt.Printf(">>Provissioning machine, Job ID %q\n", vm.JobID)

	vm.WaitVMCreation(60, 5)

	fmt.Printf("Machine created, Machine ID %q\n", vm.ID)

	fmt.Println("[Machine should be starting]")
	printInfo(vm)

	time.Sleep(30 * time.Second)
	fmt.Println("[Machine should be running]")
	printInfo(vm)

	fmt.Println(">>Starting (checking idempotency)")
	if err = vm.Start(); err != nil {
		fmt.Printf("Error starting machine: %s", err)
	}

	time.Sleep(5 * time.Second)
	fmt.Println("[Machine should be running]")
	printInfo(vm)

	fmt.Printf(">>Stopping\n")
	if err = vm.Halt(); err != nil {
		fmt.Printf("Error stopping machine: %s", err)
	}

	fmt.Println("[Machine should be stopping/running]")
	printInfo(vm)

	time.Sleep(15 * time.Second)
	fmt.Println("[Machine should be stopped]")
	printInfo(vm)

	fmt.Printf(">>Starting\n")
	if err = vm.Start(); err != nil {
		fmt.Printf("Error starting machine: %s", err)
	}
	fmt.Println("[Machine should be starting]")
	printInfo(vm)

	time.Sleep(20 * time.Second)
	fmt.Println("[Machine should be running]")
	printInfo(vm)

	fmt.Printf("Destroy ...\n")
	vm.Destroy()

	fmt.Println("[Machine should be stopping]")
	printInfo(vm)

}

func printInfo(vm exoscale.VM) {
	name := vm.GetName()

	state, err := vm.GetState()
	if err != nil {
		fmt.Printf("Error reading state: %s", err)
	}

	ips, err := vm.GetIPs()
	if err != nil {
		fmt.Printf("Error reading IPs: %s", err)
	}

	fmt.Printf("\tName %s\n\tState: %s\n\tIPs: %v\n", name, state, ips)
}

func usage() {
	fmt.Printf("Environment variables\n\tEXOSCALE_API_KEY\n\tEXOSCALE_API_SECRET\nmust be set\n")
}

func createExoscalVM(key string, secret string) exoscale.VM {
	config := exoscale.Config{
		Endpoint:  "https://api.exoscale.ch/compute",
		APIKey:    key,
		APISecret: secret,
	}

	template := exoscale.Template{
		Name:      "Linux Ubuntu 16.04 LTS 64-bit",
		StorageGB: 10,
		ZoneName:  "ch-dk-2",
	}

	service := exoscale.ServiceOffering{
		Name: exoscale.Micro,
	}

	security := []exoscale.SecurityGroup{
		{Name: "default"},
		{Name: "second sg"},
	}

	zone := exoscale.Zone{
		Name: "ch-dk-2",
	}

	vm := exoscale.VM{
		Config:          config,
		Template:        template,
		ServiceOffering: service,
		SecurityGroups:  security,
		KeypairName:     "first",
		Userdata:        "#cloud-config\nmanage_etc_hosts: true\nfqdn: new.host\n",
		Zone:            zone,
		Name:            "libratto-exoscale",
	}

	return vm
}
