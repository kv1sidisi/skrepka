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

resource "docker_image" "skrepka_backend_image" {
  name = "skrepka-backend:latest"
  build {
    context = ".."
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
    "DB_SOURCE=postgresql://${var.db_user}:${var.db_password}@${docker_container.postgres.name}:${var.db_port}/${var.db_name}?sslmode=disable"
  ]

  depends_on = [docker_container.postgres]
}

resource "docker_volume" "postgres_data" {
  name = "postgres_data"
}