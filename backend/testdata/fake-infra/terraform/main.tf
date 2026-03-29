terraform {
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
  }
}

resource "null_resource" "hello" {
  triggers = {
    always = timestamp()
  }

  provisioner "local-exec" {
    command = "echo 'Hello from Kiln!'"
  }
}

resource "null_resource" "setup" {
  triggers = {
    always = timestamp()
  }

  provisioner "local-exec" {
    command = "echo 'Setting up environment: ${var.env}'"
  }

  depends_on = [null_resource.hello]
}

resource "null_resource" "validate" {
  triggers = {
    always = timestamp()
  }

  provisioner "local-exec" {
    command = "echo 'Validation complete for ${var.env}'"
  }

  depends_on = [null_resource.setup]
}

output "environment" {
  value = var.env
}
