variable "TAG" {
    default = "latest"
}

group "default" {
    targets = ["supplier", "reactor", "grafana", "mq", "web"]
}

target "supplier" {
    context = "./supplier"
    tags = [
        "ghcr.io/chitoku-k/ejaculation-counter/supplier:latest",
        "ghcr.io/chitoku-k/ejaculation-counter/supplier:${TAG}",
    ]
}

target "reactor" {
    context = "./reactor"
    tags = [
        "ghcr.io/chitoku-k/ejaculation-counter/reactor:latest",
        "ghcr.io/chitoku-k/ejaculation-counter/reactor:${TAG}",
    ]
}

target "grafana" {
    context = "./grafana"
    tags = [
        "ghcr.io/chitoku-k/ejaculation-counter/grafana:latest",
        "ghcr.io/chitoku-k/ejaculation-counter/grafana:${TAG}",
    ]
}

target "mq" {
    context = "./mq"
    tags = [
        "ghcr.io/chitoku-k/ejaculation-counter/mq:latest",
        "ghcr.io/chitoku-k/ejaculation-counter/mq:${TAG}",
    ]
}

target "web" {
    context = "."
    tags = [
        "ghcr.io/chitoku-k/ejaculation-counter/web:latest",
        "ghcr.io/chitoku-k/ejaculation-counter/web:${TAG}",
    ]
}
