resource "kubernetes_config_map" "my_config_map" {
  metadata {
    name = "my-config-map"
  }

  data = {
    what_is_this = "data"
  }
}