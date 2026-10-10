package com.example.starterapi.adapter.out.client;

import com.sun.net.httpserver.HttpServer;
import io.quarkus.test.common.QuarkusTestResourceLifecycleManager;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.atomic.AtomicInteger;

/** Provedor ViaCEP falso (servidor HTTP do JDK) para testar o adapter de ponta a ponta. */
public class ViaCepStub implements QuarkusTestResourceLifecycleManager {

  static final AtomicInteger calls = new AtomicInteger();
  static final List<String> correlationIds = new CopyOnWriteArrayList<>();

  private HttpServer server;

  @Override
  public Map<String, String> start() {
    try {
      server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
    } catch (IOException e) {
      throw new IllegalStateException(e);
    }
    server.createContext(
        "/",
        exchange -> {
          calls.incrementAndGet();
          String id = exchange.getRequestHeaders().getFirst("X-Correlation-Id");
          if (id != null) {
            correlationIds.add(id);
          }
          String path = exchange.getRequestURI().getPath();
          int status = 200;
          String body;
          switch (path) {
            case "/01001-000/json", "/01001000/json" ->
                body =
                    """
                    {"cep":"01001-000","logradouro":"Praça da Sé","complemento":"lado ímpar",
                     "bairro":"Sé","localidade":"São Paulo","uf":"SP"}""";
            case "/99999999/json" -> body = "{\"erro\": true}";
            case "/00000000/json" -> {
              status = 400;
              body = "";
            }
            default -> {
              status = 500;
              body = "{}";
            }
          }
          byte[] bytes = body.getBytes(StandardCharsets.UTF_8);
          exchange.getResponseHeaders().add("Content-Type", "application/json");
          exchange.sendResponseHeaders(status, bytes.length == 0 ? -1 : bytes.length);
          if (bytes.length > 0) {
            exchange.getResponseBody().write(bytes);
          }
          exchange.close();
        });
    server.start();
    return Map.of(
        "quarkus.rest-client.viacep.url", "http://127.0.0.1:" + server.getAddress().getPort());
  }

  @Override
  public void stop() {
    if (server != null) {
      server.stop(0);
    }
  }
}
