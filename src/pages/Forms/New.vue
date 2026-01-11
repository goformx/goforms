<script setup lang="ts">
import { ref, computed } from "vue";
import { useForm, router } from "@inertiajs/vue3";
import DashboardLayout from "@/components/layout/DashboardLayout.vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-vue-next";

interface Props {
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = defineProps<Props>();

const form = useForm({
  title: "",
  description: "",
});

const serverError = ref<string | null>(null);
const isSubmitting = computed(() => form.processing);

function handleSubmit() {
  serverError.value = null;

  if (!form.title.trim()) {
    serverError.value = "Form title is required";
    return;
  }

  form.post("/forms", {
    onSuccess: (page) => {
      // Redirect to edit page with the new form ID
      const formId = (page.props as { formId?: string }).formId;
      if (formId) {
        router.visit(`/forms/${formId}/edit`);
      }
    },
    onError: (errors) => {
      if (errors['title']) {
        serverError.value = errors['title'];
      } else {
        serverError.value = "Failed to create form. Please try again.";
      }
    },
  });
}
</script>

<template>
  <DashboardLayout title="Create New Form" subtitle="Create a new form to collect data">
    <div class="max-w-2xl mx-auto">
      <Card class="bg-card/50 backdrop-blur-sm border-border/50">
        <CardHeader>
          <CardTitle>Form Details</CardTitle>
          <CardDescription>
            Enter the basic details for your new form. You can add fields after creating it.
          </CardDescription>
        </CardHeader>

        <form @submit.prevent="handleSubmit">
          <CardContent class="space-y-4">
            <!-- Error Alert -->
            <Alert v-if="serverError || props.flash?.error" variant="destructive"
              class="bg-destructive/15 border-destructive text-destructive">
              <AlertCircle class="h-4 w-4" />
              <AlertDescription class="font-medium">
                {{ serverError || props.flash?.error }}
              </AlertDescription>
            </Alert>

            <!-- Title Field -->
            <div class="space-y-2">
              <Label for="title">Form Title <span class="text-destructive">*</span></Label>
              <Input id="title" v-model="form.title" type="text" placeholder="Enter form title" required />
            </div>

            <!-- Description Field -->
            <div class="space-y-2">
              <Label for="description">Description</Label>
              <textarea id="description" v-model="form.description"
                class="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                placeholder="Enter form description (optional)" rows="3" />
            </div>
          </CardContent>

          <CardFooter class="flex justify-between">
            <Button type="button" variant="outline" @click="router.visit('/dashboard')">
              Cancel
            </Button>
            <Button type="submit"
              class="bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600 text-white border-0"
              :disabled="isSubmitting">
              <span v-if="isSubmitting">Creating...</span>
              <span v-else>Create Form</span>
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  </DashboardLayout>
</template>
