<script setup lang="ts">
import { Head } from "@inertiajs/vue3";
import Nav from "@/components/shared/Nav.vue";
import Footer from "@/components/shared/Footer.vue";

interface Props {
  title?: string;
}

withDefaults(defineProps<Props>(), {
  title: "GoFormX",
});
</script>

<template>
  <div class="min-h-screen flex flex-col bg-background">
    <Head :title="title" />
    
    <header>
      <Nav />
    </header>
    
    <main class="flex-1">
      <slot />
    </main>
    
    <Footer />
  </div>
</template>
