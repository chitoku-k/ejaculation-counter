group "default" {
    targets = ["supplier", "reactor", "grafana", "mq", "web"]
}

target "supplier" {
    context = "./supplier"
    target = "production"
}

target "reactor" {
    context = "./reactor"
    target = "production"
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
