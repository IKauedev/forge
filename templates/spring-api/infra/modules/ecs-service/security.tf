resource "aws_security_group" "alb" {
  name        = "${var.name}-alb"
  description = "ALB ${var.name}: entrada publica HTTP/HTTPS"
  vpc_id      = var.vpc_id

  tags = {
    Name = "${var.name}-alb"
  }
}

# O ALB é público de propósito: é a porta de entrada da API.
resource "aws_vpc_security_group_ingress_rule" "alb_http" {
  security_group_id = aws_security_group.alb.id
  description       = "HTTP publico"
  ip_protocol       = "tcp"
  from_port         = 80
  to_port           = 80
  cidr_ipv4         = "0.0.0.0/0"
}

resource "aws_vpc_security_group_ingress_rule" "alb_https" {
  security_group_id = aws_security_group.alb.id
  description       = "HTTPS publico"
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
  cidr_ipv4         = "0.0.0.0/0"
}

resource "aws_vpc_security_group_egress_rule" "alb_to_service" {
  security_group_id            = aws_security_group.alb.id
  description                  = "Trafego para as tasks"
  referenced_security_group_id = aws_security_group.service.id
  ip_protocol                  = "tcp"
  from_port                    = var.container_port
  to_port                      = var.container_port
}

resource "aws_security_group" "service" {
  name        = "${var.name}-service"
  description = "Tasks ${var.name}: aceita apenas o ALB"
  vpc_id      = var.vpc_id

  tags = {
    Name = "${var.name}-service"
  }
}

resource "aws_vpc_security_group_ingress_rule" "service_from_alb" {
  security_group_id            = aws_security_group.service.id
  description                  = "Trafego vindo do ALB"
  referenced_security_group_id = aws_security_group.alb.id
  ip_protocol                  = "tcp"
  from_port                    = var.container_port
  to_port                      = var.container_port
}

# As tasks precisam sair para ECR, Secrets Manager, CloudWatch, o banco e APIs externas (ViaCEP).
resource "aws_vpc_security_group_egress_rule" "service_all" {
  security_group_id = aws_security_group.service.id
  description       = "Saida das tasks"
  ip_protocol       = "-1"
  cidr_ipv4         = "0.0.0.0/0"
}
