<script setup lang="ts">
const currentYear = new Date().getFullYear();
</script>

<template>
  <footer class="border-t py-6 md:py-0">
    <div class="container flex flex-col items-center justify-between gap-4 md:h-16 md:flex-row">
      <p class="text-sm text-muted-foreground">
        &copy; {{ currentYear }} GoFormX. All rights reserved.
      </p>
      <nav class="flex items-center space-x-4 text-sm text-muted-foreground">
        <a
          href="https://github.com/goformx/goforms"
          target="_blank"
          rel="noopener noreferrer"
          class="hover:text-foreground transition-colors"
        >
          GitHub
        </a>
        <a
          href="https://github.com/goformx/goforms/blob/main/LICENSE"
          target="_blank"
          rel="noopener noreferrer"
          class="hover:text-foreground transition-colors"
        >
          License
        </a>
      </nav>
    </div>
  </footer>
</template>
