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
    <section class="relative py-20 md:py-32 min-h-[calc(100vh-3.5rem)] flex items-center overflow-hidden">
      <!-- Gradient background -->
      <div class="absolute inset-0 bg-gradient-to-b from-background via-background to-background/80" />
      
      <!-- Radial accent glows -->
      <div class="absolute inset-0 overflow-hidden">
        <div class="absolute top-[20%] left-[10%] w-[500px] h-[500px] bg-indigo-500/15 rounded-full blur-3xl" />
        <div class="absolute bottom-[10%] right-[10%] w-[400px] h-[400px] bg-purple-500/15 rounded-full blur-3xl" />
        <div class="absolute top-[50%] right-[30%] w-[300px] h-[300px] bg-violet-500/10 rounded-full blur-3xl" />
      </div>
      
      <!-- Content -->
      <div class="container relative z-10">
        <div class="flex flex-col items-center text-center">
          <h1 class="text-4xl font-bold tracking-tight sm:text-5xl md:text-6xl lg:text-7xl">
            <span class="bg-gradient-to-r from-white via-white to-white/80 bg-clip-text text-transparent">Your Forms,</span>
            <br />
            <span class="bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent italic">Our Backend</span>
          </h1>
          <p class="mt-6 max-w-[42rem] text-lg text-muted-foreground sm:text-xl">
            Use our free hosted service or deploy your own. A self-hosted form backend built with Go.
          </p>
          <p class="mt-2 text-sm text-muted-foreground/70">
            No dependencies, just Go.
          </p>
          <div class="mt-8 flex flex-col gap-4 sm:flex-row">
            <Button size="lg" class="bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600 text-white border-0" as-child>
              <Link href="/signup">Try demo</Link>
            </Button>
            <Button size="lg" variant="outline" class="border-border/50 bg-background/50 backdrop-blur hover:bg-background/80" as-child>
              <Link href="https://github.com/goformx/goforms" target="_blank">View on GitHub</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="relative py-20">
      <!-- Subtle gradient overlay -->
      <div class="absolute inset-0 bg-gradient-to-b from-transparent via-muted/30 to-transparent" />
      
      <div class="container relative z-10">
        <div class="text-center mb-12">
          <h2 class="text-3xl font-bold tracking-tight sm:text-4xl">
            Everything You Need
          </h2>
          <p class="mt-4 text-lg text-muted-foreground">
            Powerful features to help you build and manage forms at scale.
          </p>
        </div>
        <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          <Card 
            v-for="feature in features" 
            :key="feature.title"
            class="bg-card/50 backdrop-blur-sm border-border/50 hover:bg-card/70 hover:border-border transition-all duration-300"
          >
            <CardHeader>
              <div class="mb-2 flex h-12 w-12 items-center justify-center rounded-lg bg-gradient-to-br from-indigo-500/20 to-purple-500/20">
                <component :is="feature.icon" class="h-6 w-6 text-indigo-400" />
              </div>
              <CardTitle>{{ feature.title }}</CardTitle>
              <CardDescription>{{ feature.description }}</CardDescription>
            </CardHeader>
          </Card>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="relative py-20">
      <!-- Accent glow -->
      <div class="absolute inset-0 overflow-hidden">
        <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[300px] bg-indigo-500/10 rounded-full blur-3xl" />
      </div>
      
      <div class="container relative z-10">
        <div class="flex flex-col items-center text-center">
          <h2 class="text-3xl font-bold tracking-tight sm:text-4xl">
            Ready to Get Started?
          </h2>
          <p class="mt-4 max-w-[42rem] text-lg text-muted-foreground">
            Create your first form in minutes. No credit card required.
          </p>
          <div class="mt-8">
            <Button size="lg" class="bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600 text-white border-0" as-child>
              <Link href="/signup">Create Your Account</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  </GuestLayout>
</template>
