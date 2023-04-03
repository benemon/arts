# ARTS - Ansible Run Task Shim

# Intro

This is a spike project to look at integration between HashiCorp Terraform Cloud / Enterprise and Ansible Automation Platform / AWX using the enterprise Terraform feature of 'Run Tasks'.

A Run Task is a way of integrating an enterprise Terraform platform with 3rd party applications and services.

In lieu of actual Run Task support in Ansible Automation Platform / AWX, the goal is to create a simple shim that will act as an intermediary between the two platforms.

## Usage

This can either be built locally using the go compiler, or containerised for deployment elsewhere


```bash
oc new-build --binary --name=arts
oc start-build arts --from-dir=build/
```
