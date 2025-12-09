group "default" {
    targets = ["cloudflare-exporter"]
}

target "cloudflare-exporter" {
    context = "."
}
