# Archived deployment files

These Kops manifests and scripts describe an earlier deployment and are kept for reference only. They are not the current installation path and have not been validated for current production.

Do not run `setup-x86-only.sh` or bulk-apply this directory. Old ports, replica counts, controllers, public addresses, and cost/availability claims are obsolete. In particular, the dashboard-admin manifest grants broad administrative access and must not be enabled as part of the new setup.

Use the root deployment guide, Terraform environments, and Helm chart instead. The pre-existing ignored `../manifests/xemapp-secrets.yaml` was deliberately left in place; its contents are not migrated, read or committed by this update. Import required settings through the current secret-management workflow, never by committing that file.
