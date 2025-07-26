terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
      version = "~> 3.0.2"
    }
  }
}

provider "docker" {
  host = "ssh://${var.server_user}@${var.server_ip}:${var.server_port}"
}

locals {
  source_files = setunion(
    fileset("${path.module}/..", "**/*.go"),
    fileset("${path.module}/..", "go.mod"),
    fileset("${path.module}/..", "go.sum")
  )

  source_hash = jsonencode([
    for f in local.source_files : filesha256("${path.module}/../${f}")
  ])
}

resource "docker_image" "skrepka_backend_image" {
  name = "skrepka-backend:latest"
  build {
    context = ".."
  }

  triggers = {
    source_code_hash = local.source_hash
  }
}

resource "docker_network" "skrepka_network" {
  name = "skrepka-network"
}

resource "docker_container" "postgres" {
  image = "postgres:16"
  name  = "skrepka-postgres"

  ports {
    internal = 5432
    external = var.db_port
  }

  networks_advanced {
    name = docker_network.skrepka_network.name
  }

  env = [
    "POSTGRES_USER=${var.db_user}",
    "POSTGRES_PASSWORD=${var.db_password}",
    "POSTGRES_DB=${var.db_name}"
  ]

  volumes {
    volume_name = "postgres_data"
    container_path = "/var/lib/postgresql/data"
  }
}

resource "docker_container" "skrepka_backend" {
  image = docker_image.skrepka_backend_image.name
  name  = "skrepka-backend"

  ports {
    internal = 4000
    external = var.backend_port
  }

  networks_advanced {
    name = docker_network.skrepka_network.name
  }

  env = [
    "CONFIG_PATH=/app/configs/config.yml",
    "DB_HOST=${docker_container.postgres.name}",
    "DB_PORT=5432",
    "DB_USER=${var.db_user}",
    "DB_PASSWORD=${var.db_password}",
    "DB_NAME=${var.db_name}",
    "JWT_SECRET=${var.jwt_secret}",
    "GOOGLE_CLIENT_ID=${var.google_client_id}"
  ]


  depends_on = [docker_container.postgres]
}

resource "docker_volume" "postgres_data" {
  name = "postgres_data"
}