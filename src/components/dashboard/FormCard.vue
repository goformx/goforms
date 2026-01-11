<script setup lang="ts">
import { computed } from "vue";
import { Link, router } from "@inertiajs/vue3";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Eye, Pencil, ListChecks, MoreVertical, Copy, Download, Archive, Trash2 } from "lucide-vue-next";

interface Form {
  id: string;
  title: string;
  description: string;
  status: "draft" | "published" | "archived";
  createdAt: string;
  updatedAt: string;
}

interface Props {
  form: Form;
}

interface Emits {
  (e: "duplicate", formId: string): void;
  (e: "export", formId: string): void;
  (e: "archive", formId: string): void;
  (e: "delete", formId: string): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

// Badge variant based on status
const statusVariant = computed(() => {
  switch (props.form.status) {
    case "published":
      return "default";
    case "draft":
      return "secondary";
    case "archived":
      return "outline";
    default:
      return "secondary";
  }
});

// Format relative time
function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) return "just now";
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
  if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)}d ago`;
  if (diffInSeconds < 2592000) return `${Math.floor(diffInSeconds / 604800)}w ago`;
  if (diffInSeconds < 31536000) return `${Math.floor(diffInSeconds / 2592000)}mo ago`;
  return `${Math.floor(diffInSeconds / 31536000)}y ago`;
}

// Navigation actions
function editForm() {
  router.visit(`/forms/${props.form.id}/edit`);
}

function previewForm() {
  router.visit(`/forms/${props.form.id}/preview`);
}

function viewSubmissions() {
  router.visit(`/forms/${props.form.id}/submissions`);
}

// Dropdown actions
function duplicateForm() {
  emit("duplicate", props.form.id);
}

function exportForm() {
  emit("export", props.form.id);
}

function archiveForm() {
  emit("archive", props.form.id);
}

function deleteForm() {
  emit("delete", props.form.id);
}
</script>

<template>
  <Card
    class="form-card group hover:shadow-lg transition-all duration-200 bg-card/50 backdrop-blur-sm border-border/50 hover:bg-card/70 hover:border-border">
    <CardHeader>
      <div class="flex items-start justify-between gap-2">
        <div class="flex-1 min-w-0">
          <CardTitle class="text-lg truncate mb-2">
            {{ props.form.title }}
          </CardTitle>
          <div class="flex items-center gap-2">
            <Badge :variant="statusVariant" class="capitalize">
              {{ props.form.status }}
            </Badge>
            <span class="text-xs text-muted-foreground">
              Updated {{ formatRelativeTime(props.form.updatedAt) }}
            </span>
          </div>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="ghost" size="icon" class="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity">
              <MoreVertical class="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem @click="duplicateForm">
              <Copy class="mr-2 h-4 w-4" />
              Duplicate
            </DropdownMenuItem>
            <DropdownMenuItem @click="exportForm">
              <Download class="mr-2 h-4 w-4" />
              Export
            </DropdownMenuItem>
            <DropdownMenuItem @click="archiveForm">
              <Archive class="mr-2 h-4 w-4" />
              Archive
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem @click="deleteForm" class="text-destructive focus:text-destructive">
              <Trash2 class="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </CardHeader>

    <CardContent>
      <p class="text-sm text-muted-foreground line-clamp-2 min-h-[2.5rem]">
        {{ props.form.description || "No description" }}
      </p>

      <!-- Stats placeholder - will be populated when backend provides stats -->
      <div class="flex items-center gap-4 mt-4 text-xs text-muted-foreground">
        <div class="flex items-center gap-1">
          <Eye class="h-3.5 w-3.5" />
          <span>0 views</span>
        </div>
        <div class="flex items-center gap-1">
          <ListChecks class="h-3.5 w-3.5" />
          <span>0 submissions</span>
        </div>
      </div>
    </CardContent>

    <CardFooter class="gap-2">
      <Button variant="outline" size="sm" @click="previewForm" class="flex-1">
        <Eye class="mr-2 h-4 w-4" />
        Preview
      </Button>
      <Button size="sm" @click="editForm" class="flex-1">
        <Pencil class="mr-2 h-4 w-4" />
        Edit
      </Button>
    </CardFooter>
  </Card>
</template>

<style scoped>
.form-card {
  @apply cursor-pointer;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
