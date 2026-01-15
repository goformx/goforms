<script setup lang="ts">
import { ref, computed } from "vue";
import { Link, router } from "@inertiajs/vue3";
import DashboardLayout from "@/components/layout/DashboardLayout.vue";
import FormCard from "@/components/dashboard/FormCard.vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, Search, SlidersHorizontal } from "lucide-vue-next";
import { toast } from "vue-sonner";

interface Form {
  id: string;
  title: string;
  description: string;
  status: "draft" | "published" | "archived";
  createdAt: string;
  updatedAt: string;
}

interface Props {
  forms: Form[];
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = defineProps<Props>();

// Search and filter state
const searchQuery = ref("");
const statusFilter = ref<"all" | "draft" | "published" | "archived">("all");
const sortBy = ref<"updated" | "created" | "title">("updated");

// Computed filtered and sorted forms
const filteredForms = computed(() => {
  let filtered = props.forms ?? [];

  // Apply search filter
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(
      (form) =>
        form.title.toLowerCase().includes(query) ||
        form.description.toLowerCase().includes(query)
    );
  }

  // Apply status filter
  if (statusFilter.value !== "all") {
    filtered = filtered.filter((form) => form.status === statusFilter.value);
  }

  // Apply sorting
  filtered = [...filtered].sort((a, b) => {
    switch (sortBy.value) {
      case "title":
        return a.title.localeCompare(b.title);
      case "created":
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
      case "updated":
      default:
        return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
    }
  });

  return filtered;
});

const hasForms = computed(() => props.forms.length > 0);
const hasFilteredForms = computed(() => filteredForms.value.length > 0);

// Stats
const stats = computed(() => ({
  total: props.forms?.length ?? 0,
  published: props.forms?.filter((f) => f.status === "published").length ?? 0,
  draft: props.forms?.filter((f) => f.status === "draft").length ?? 0,
  archived: props.forms?.filter((f) => f.status === "archived").length ?? 0,
}));

// Form actions
function duplicateForm(_formId: string) {
  toast.info("Duplicate functionality coming soon!");
}

async function exportForm(formId: string) {
  try {
    const response = await fetch(`/api/v1/forms/${formId}/schema`);
    if (!response.ok) throw new Error("Failed to fetch form schema");

    const data = await response.json();
    const schema = data.success ? data.data : data;

    // Download as JSON file
    const blob = new Blob([JSON.stringify(schema, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `form-${formId}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    toast.success("Form exported successfully");
  } catch (err) {
    toast.error("Failed to export form");
    console.error(err);
  }
}

function archiveForm(_formId: string) {
  toast.info("Archive functionality coming soon!");
}

async function deleteForm(formId: string) {
  if (!confirm("Are you sure you want to delete this form? This action cannot be undone.")) {
    return;
  }

  router.delete(`/forms/${formId}`, {
    preserveScroll: true,
    onSuccess: () => {
      toast.success("Form deleted successfully");
    },
    onError: () => {
      toast.error("Failed to delete form");
    },
  });
}

const clearFilters = () => {
  searchQuery.value = "";
  statusFilter.value = "all";
};

// Show flash messages
if (props.flash?.success) {
  toast.success(props.flash.success);
}
if (props.flash?.error) {
  toast.error(props.flash.error);
}
</script>

<template>
  <DashboardLayout title="Your Forms" subtitle="Manage and create forms">
    <template #actions>
      <Button
        class="bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600 text-white border-0"
        as-child>
        <Link href="/forms/new">
          <Plus class="mr-2 h-4 w-4" />
          New Form
        </Link>
      </Button>
    </template>

    <div class="space-y-6">
      <!-- Stats Overview -->
      <div v-if="hasForms" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card class="bg-card/50 backdrop-blur-sm border-border/50">
          <CardContent class="pt-6">
            <div class="text-2xl font-bold">{{ stats.total }}</div>
            <p class="text-xs text-muted-foreground">Total Forms</p>
          </CardContent>
        </Card>
        <Card class="bg-card/50 backdrop-blur-sm border-border/50">
          <CardContent class="pt-6">
            <div class="text-2xl font-bold text-green-400">{{ stats.published }}</div>
            <p class="text-xs text-muted-foreground">Published</p>
          </CardContent>
        </Card>
        <Card class="bg-card/50 backdrop-blur-sm border-border/50">
          <CardContent class="pt-6">
            <div class="text-2xl font-bold text-yellow-400">{{ stats.draft }}</div>
            <p class="text-xs text-muted-foreground">Drafts</p>
          </CardContent>
        </Card>
        <Card class="bg-card/50 backdrop-blur-sm border-border/50">
          <CardContent class="pt-6">
            <div class="text-2xl font-bold text-gray-400">{{ stats.archived }}</div>
            <p class="text-xs text-muted-foreground">Archived</p>
          </CardContent>
        </Card>
      </div>

      <!-- Search and Filters -->
      <div v-if="hasForms" class="flex flex-col sm:flex-row gap-4">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input v-model="searchQuery" type="search" placeholder="Search forms by title or description..."
            class="pl-9" />
        </div>
        <div class="flex gap-2">
          <Select v-model="statusFilter">
            <SelectTrigger class="w-[140px]">
              <SelectValue placeholder="All Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="draft">Draft</SelectItem>
              <SelectItem value="published">Published</SelectItem>
              <SelectItem value="archived">Archived</SelectItem>
            </SelectContent>
          </Select>
          <Select v-model="sortBy">
            <SelectTrigger class="w-[140px]">
              <SelectValue placeholder="Sort by" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="updated">Last Updated</SelectItem>
              <SelectItem value="created">Date Created</SelectItem>
              <SelectItem value="title">Title (A-Z)</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <!-- Empty State (No Forms) -->
      <Card v-if="!hasForms" class="bg-card/50 backdrop-blur-sm border-border/50">
        <CardContent class="flex flex-col items-center justify-center py-16">
          <div class="text-center max-w-sm">
            <div
              class="mx-auto mb-4 h-12 w-12 rounded-full bg-gradient-to-br from-indigo-500/20 to-purple-500/20 flex items-center justify-center">
              <Plus class="h-6 w-6 text-indigo-400" />
            </div>
            <h3 class="text-lg font-semibold mb-2">No forms yet</h3>
            <p class="text-muted-foreground mb-6">
              Get started by creating your first form. You can collect responses, analyze data, and share your forms
              with
              anyone.
            </p>
            <Button
              class="bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600 text-white border-0"
              as-child>
              <Link href="/forms/new">
                <Plus class="mr-2 h-4 w-4" />
                Create Your First Form
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- Empty State (No Results) -->
      <Card v-else-if="!hasFilteredForms" class="bg-card/50 backdrop-blur-sm border-border/50">
        <CardContent class="flex flex-col items-center justify-center py-16">
          <div class="text-center max-w-sm">
            <SlidersHorizontal class="mx-auto h-12 w-12 text-muted-foreground/50 mb-4" />
            <h3 class="text-lg font-semibold mb-2">No forms found</h3>
            <p class="text-muted-foreground mb-6">
              No forms match your current filters. Try adjusting your search or filter criteria.
            </p>
            <Button variant="outline" class="border-border/50" @click="clearFilters">
              Clear Filters
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- Forms Grid -->
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <FormCard v-for="form in filteredForms" :key="form.id" :form="form" @duplicate="duplicateForm"
          @export="exportForm" @archive="archiveForm" @delete="deleteForm" />
      </div>
    </div>
  </DashboardLayout>
</template>
