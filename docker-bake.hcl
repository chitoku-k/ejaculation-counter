group "default" {
    targets = ["supplier", "reactor", "grafana", "mq", "web"]
}

target "supplier" {
    context = "./supplier"
}

target "reactor" {
    context = "./reactor"
}

target "grafana" {
    context = "./grafana"
}

target "mq" {
    context = "./mq"
}

target "web" {
    context = "./web"
}
