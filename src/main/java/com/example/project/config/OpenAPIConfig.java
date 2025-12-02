package com.example.project.config;

import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Contact;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.info.License;
import io.swagger.v3.oas.models.servers.Server;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.List;

@Configuration
public class OpenAPIConfig {

    @Bean
    public OpenAPI myOpenAPI() {
        Server devServer = new Server();
        devServer.setUrl("http://localhost:8080");
        devServer.setDescription("Сервер для разработки");

        Server prodServer = new Server();
        prodServer.setUrl("https://fitness-club-production.com");
        prodServer.setDescription("Продакшен сервер");

        Contact contact = new Contact();
        contact.setEmail("support@fitnessclub.com");
        contact.setName("Fitness Club Support");
        contact.setUrl("https://fitnessclub.com");

        License mitLicense = new License()
                .name("MIT License")
                .url("https://choosealicense.com/licenses/mit/");

        Info info = new Info()
                .title("Fitness Club Management API")
                .version("1.0")
                .contact(contact)
                .description("API для системы управления фитнес-клубом")
                .license(mitLicense);

        return new OpenAPI()
                .info(info)
                .servers(List.of(devServer, prodServer));
    }
}