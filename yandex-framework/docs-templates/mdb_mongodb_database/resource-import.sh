# The resource can be imported by using their resource ID.
# For getting a resource ID you can use Yandex Cloud Web UI or YC CLI.

# A MongoDB Database can be imported using the following format:
terraform import yandex_mdb_mongodb_database.foo {cluster_id}:{database_name}
