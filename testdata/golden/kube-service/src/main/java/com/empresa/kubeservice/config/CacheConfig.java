package com.empresa.kubeservice.config;

import org.springframework.cache.annotation.EnableCaching;
import org.springframework.context.annotation.Configuration;

/** Habilita @Cacheable. O tipo do cache vem de spring.cache.type (caffeine | redis). */
@Configuration(proxyBeanMethods = false)
@EnableCaching
public class CacheConfig {}
