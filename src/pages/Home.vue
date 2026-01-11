<script setup lang="ts">
import { Link } from "@inertiajs/vue3";
import GuestLayout from "@/components/layout/GuestLayout.vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { FileText, Zap, Shield, Code, Database, Globe } from "lucide-vue-next";

const features = [
  {
    icon: FileText,
    title: "Visual Form Builder",
    description: "Create forms visually with drag-and-drop. No coding required.",
  },
  {
    icon: Zap,
    title: "Real-time Validation",
    description: "Instant feedback with Zod-powered validation schemas.",
  },
  {
    icon: Shield,
    title: "Secure by Default",
    description: "Built-in CSRF protection, rate limiting, and input sanitization.",
  },
  {
    icon: Code,
    title: "API-First Design",
    description: "RESTful API for embedding forms anywhere on the web.",
  },
  {
    icon: Database,
    title: "Self-Hosted",
    description: "Your data stays on your servers. Full control and privacy.",
  },
  {
    icon: Globe,
    title: "CORS Support",
    description: "Embed forms on any domain with configurable CORS policies.",
  },
];
</script>

<template>
  <GuestLayout title="Home">
    <!-- Hero Section -->
    <section class="py-20 md:py-32">
      <div class="container">
        <div class="flex flex-col items-center text-center">
          <h1 class="text-4xl font-bold tracking-tight sm:text-5xl md:text-6xl lg:text-7xl">
            Build Beautiful Forms
            <span class="text-primary">Effortlessly</span>
          </h1>
          <p class="mt-6 max-w-[42rem] text-lg text-muted-foreground sm:text-xl">
            A self-hosted form backend service built with Go. Create, manage, and collect form submissions with a modern, secure platform.
          </p>
          <div class="mt-8 flex flex-col gap-4 sm:flex-row">
            <Button size="lg" as-child>
              <Link href="/signup">Get Started</Link>
            </Button>
            <Button size="lg" variant="outline" as-child>
              <Link href="/demo">View Demo</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="py-20 bg-muted/50">
      <div class="container">
        <div class="text-center mb-12">
          <h2 class="text-3xl font-bold tracking-tight sm:text-4xl">
            Everything You Need
          </h2>
          <p class="mt-4 text-lg text-muted-foreground">
            Powerful features to help you build and manage forms at scale.
          </p>
        </div>
        <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          <Card v-for="feature in features" :key="feature.title">
            <CardHeader>
              <div class="mb-2 flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                <component :is="feature.icon" class="h-6 w-6 text-primary" />
              </div>
              <CardTitle>{{ feature.title }}</CardTitle>
              <CardDescription>{{ feature.description }}</CardDescription>
            </CardHeader>
          </Card>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="py-20">
      <div class="container">
        <div class="flex flex-col items-center text-center">
          <h2 class="text-3xl font-bold tracking-tight sm:text-4xl">
            Ready to Get Started?
          </h2>
          <p class="mt-4 max-w-[42rem] text-lg text-muted-foreground">
            Create your first form in minutes. No credit card required.
          </p>
          <div class="mt-8">
            <Button size="lg" as-child>
              <Link href="/signup">Create Your Account</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  </GuestLayout>
</template>
