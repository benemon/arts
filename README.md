# ARTS - Ansible Run Task Shim

# Intro

This is a spike project to look at integration between HashiCorp Terraform Cloud / Enterprise and Ansible Automation Platform / AWX using the enterprise Terraform feature of 'Run Tasks'.

A Run Task is a way of integrating an enterprise Terraform platform with 3rd party applications and services.

In lieu of [actual Run Task support](https://github.com/ansible/awx/issues/13785) in Ansible Automation Platform (AAP) / AWX, the goal is to create a simple shim that will act as an intermediary between the two platforms.

## Usage

ARTs currently provides a mechanism to trigger the following actions in AAP/AWX:

* Job Template Launching - Trigger AAP/AWX Job Templates. Note that the success criteria here is that we were able to succesfully trigger the JT, not that the JT itself completed successfully.

* Workflow Job Template Launching - Trigger more compliex AAP/AWX Workflow Job Templates. As with Job Templates, note that the success criteria here is that we were able to succesfully trigger the Workflow JT, not that the Workflow JT itself completed successfully.

* Inventory Creation - An Inventory will be created based on the Workspace Name. This wil become more useful if / when TFE/TFC support post-apply Run Tasks as we will be able to pre-populatre Ansible Inventories with IPs and Hostnames generated directly by a Terraform Apply, or hand crafted in Terraform Outputs.

## Configuration

### Build
This can either be built locally using the go compiler, or containerised for deployment elsewhere. 

To deploy onto OKD/OpenShift Container Platform

```bash
$ oc new-project arts
$ oc new-build --binary --name=arts
$ ./build.sh
$ oc apply -f deployment/
```

### ARTS
The only configuration required for ARTs is the resolvable FQDN name of the Ansible Automation Platform (AAP) / AWX Controller, and the initial credentials with which to authenticate against it.

These are supplied as the following Environment Variables:

```
ARTS_ANSIBLE_HOST - Controller FQDN
ARTS_ANSIBLE_USER - Controller Credential Username
ARTS_ANSIBLE_PASSWORD - Controller Credential Password
```

### Terraform Cloud / Enterprise

ARTs needs to be configured as a Run Task within your Organisation Settings. The structure of the ARTs Run Tasks follows a very specific pattern:

```
https://{fqdn of arts}/public/{job/workflow/inventory}/{identifier}
```

where the identifier can be one of:

* Job Template ID for the `job` endpoint e.g. `https://my-arts-shim.onmi.cloud/public/job/1`
* Workflow Job Template ID for the `workflow` endpoint e.g. `https://my-arts-shim.onmi.cloud/public/workflow/8`
* Organisation ID for the `inventory` endpoint e.g. `https://my-arts-shim.onmi.cloud/public/inventory/1`

This obviously means that to chain different AAP/AWX triggers, you must create different Run Tasks for each relevant Job Template, Workflow Job Template, or Inventory creation you wish to trigger.

The Details link from the Run Task in TFE/TFC will take you to the artifact in AAP/AWX. From there, if you have valid credentials for that platform, you'll be able to view the status of triggered process.

### Authentication

On the subject of authentication, ARTs will generate an OAuth Token from AAP/AWX for each request based on the supplied credentials, and then revoke it irrespective of the outcome of that request.

It does not currently support the use of an HMAC Key to verify the authenticity of Run Task Requests from the enterprise Terraform Platform.

### Screenshots
![Run Task - Setup Screenshot](images/setup.png)

![Run Task - Inventory Success Screenshot](images/inventory-success.png)

![Run Task - Inventory Failure Screenshot](images/inventory-fail.png)

![Run Task - Job Template Success Screenshot](images/jt-success.png)

![Run Task - Job Template Success Screenshot](images/wfjt-success.png)