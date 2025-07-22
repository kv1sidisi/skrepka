
# Server
variable "server_ip" {
  type = string
  default = "kvisidisi.ru"
  description = "Server IP-address"
}

variable "server_user" {
  type = string
  default = "kvisidisi"
  description = "SSH-connection user"
}

variable "server_port" {
  type = number
  default = 52
  description = "Server Port"
}


# Skrepka backend
variable "backend_port" {
  type = number
  default = 4000
  description = "External port for skrepka backend"
}


# Postgres
variable "db_user" {
  type = string
  default = "skrepka_user"
  description = "Skrepka postgres database user name"

}

variable "db_password" {
  type = string
  default = "skrepka_password"
  description = "Skrepka postgres database password"

}

variable "db_name" {
  type = string
  default = "skrepka_db"
  description = "Skrepka postgres database name"
}

variable "db_port" {
  type = number
  default = 5432
  description = "Skrepka postgres database external port"
}