variable "GO_VERSION" { default = "1.22" }
variable "ALPINE_VERSION" { default = "3.19" }

target "docker-metadata-action" {}
target "github-metadata-action" {}

target "default" {
    inherits = [ "prometheus-config-provider" ]
    platforms = [
        "linux/amd64",
        "linux/arm64"
    ]
}

target "local" {
    inherits = [ "prometheus-config-provider" ]
    tags = [ "swarmlibs/prometheus-config-provider:local" ]
}

target "prometheus-config-provider" {
    context = "."
    dockerfile = "Dockerfile"
    inherits = [
        "docker-metadata-action",
        "github-metadata-action",
    ]
    args = {
        GO_VERSION = "${GO_VERSION}"
        ALPINE_VERSION = "${ALPINE_VERSION}"
    }
}
