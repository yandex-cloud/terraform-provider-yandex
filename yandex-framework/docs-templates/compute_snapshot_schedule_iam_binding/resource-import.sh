# The resource can be imported by using their resource ID.
# For getting a resource ID you can use Yandex Cloud Web UI or YC CLI.

# IAM binding imports use space-delimited identifiers;
# first the resource in question and then the role. 

# These bindings can be imported using the snapshot_schedule_id and role, e.g.
terraform import yandex_compute_snapshot_schedule_iam_binding.editor "snapshot_schedule_id editor"
