group "default" {
    targets = ["supplier", "reactor", "mq", "web"]
}

target "supplier" {
    context = "./supplier"
    tags = ["container.chitoku.jp/chitoku-k/ejaculation-counter/supplier"]
}

target "reactor" {
    context = "./reactor"
    tags = ["container.chitoku.jp/chitoku-k/ejaculation-counter/reactor"]
}

target "grafana" {
    context = "./grafana"
    tags = ["container.chitoku.jp/chitoku-k/ejaculation-counter/grafana"]
}

target "mq" {
    context = "./mq"
    tags = ["container.chitoku.jp/chitoku-k/ejaculation-counter/mq"]
}

target "web" {
    context = "."
    tags = ["container.chitoku.jp/chitoku-k/ejaculation-counter/web"]
}
