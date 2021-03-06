resource "skysql_service" "wat" {
  release_version = "MariaDB Enterprise Server 10.6.4-1"
  topology        = "Single Node Transactions"
  size            = "Sky-2x4"
  tx_storage      = "100"
  name            = "standalone-example"
  region          = "ca-central-1"
  cloud_provider  = "Amazon AWS"
  replicas        = "0"
  monitor         = "false"
  volume_iops     = "100"
  maxscale_proxy  = "false"
  tier            = "Foundation"
}
