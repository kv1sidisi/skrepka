
# Server
variable "server_ip" {
  type = string
  description = "Server IP-address"
  # No default, this must be provided.
}

variable "server_user" {
  type = string
  description = "SSH-connection user"
  # No default, this must be provided.
}

variable "server_port" {
  type = number
  description = "Server Port"
  # No default, this must be provided.
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
  description = "Skrepka postgres database password"
  sensitive   = true
}

variable "db_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "skrepka_db"
}

variable "jwt_secret" {
  description = "Secret key for signing JWTs"
  type        = string
  sensitive   = true
}

variable "google_client_id" {
  description = "Google Client ID for OIDC"
  type        = string
  sensitive   = true
}

variable "db_port" {
  type = number
  default = 5432
  description = "Skrepka postgres database external port"
}