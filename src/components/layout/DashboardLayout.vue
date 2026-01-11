<script setup lang="ts">
import { Head } from "@inertiajs/vue3";
import Nav from "@/components/shared/Nav.vue";
import Footer from "@/components/shared/Footer.vue";
import DashboardHeader from "@/components/shared/DashboardHeader.vue";

interface Props {
  title?: string;
  subtitle?: string;
}

withDefaults(defineProps<Props>(), {
  title: "Dashboard",
  subtitle: "",
});
</script>

<template>
  <div class="min-h-screen flex flex-col bg-background">
    <Head :title="title" />
    
    <header>
      <Nav />
    </header>
    
    <main class="flex-1">
      <div class="container py-6">
        <DashboardHeader :title="title" :subtitle="subtitle">
          <template #actions>
            <slot name="actions" />
          </template>
        </DashboardHeader>
        
        <div class="mt-6">
          <slot />
        </div>
      </div>
    </main>
    
    <Footer />
  </div>
</template>
