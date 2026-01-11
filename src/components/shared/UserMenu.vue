<script setup lang="ts">
import { computed } from "vue";
import { Link, router } from "@inertiajs/vue3";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { User, Settings, LayoutDashboard, LogOut } from "lucide-vue-next";

interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role: string;
}

interface Props {
  user: User;
}

const props = defineProps<Props>();

const displayName = computed(() => {
  if (props.user.firstName) {
    return `${props.user.firstName} ${props.user.lastName}`.trim();
  }
  return "User";
});

const initials = computed(() => {
  if (props.user.firstName) {
    return props.user.firstName.charAt(0).toUpperCase();
  }
  return "U";
});

function logout() {
  router.post("/logout");
}
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button variant="ghost" class="relative h-8 w-8 rounded-full">
        <span
          class="flex h-8 w-8 shrink-0 overflow-hidden rounded-full bg-primary text-primary-foreground items-center justify-center text-sm font-medium"
        >
          {{ initials }}
        </span>
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent class="w-56" align="end">
      <DropdownMenuLabel class="font-normal">
        <div class="flex flex-col space-y-1">
          <p class="text-sm font-medium leading-none">{{ displayName }}</p>
          <p class="text-xs leading-none text-muted-foreground">
            {{ user.email }}
          </p>
        </div>
      </DropdownMenuLabel>
      <DropdownMenuSeparator />
      <DropdownMenuItem as-child>
        <Link href="/dashboard" class="flex items-center cursor-pointer">
          <LayoutDashboard class="mr-2 h-4 w-4" />
          <span>Dashboard</span>
        </Link>
      </DropdownMenuItem>
      <DropdownMenuItem as-child>
        <Link href="/profile" class="flex items-center cursor-pointer">
          <User class="mr-2 h-4 w-4" />
          <span>Profile</span>
        </Link>
      </DropdownMenuItem>
      <DropdownMenuItem as-child>
        <Link href="/settings" class="flex items-center cursor-pointer">
          <Settings class="mr-2 h-4 w-4" />
          <span>Settings</span>
        </Link>
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <DropdownMenuItem @click="logout" class="cursor-pointer">
        <LogOut class="mr-2 h-4 w-4" />
        <span>Log out</span>
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
