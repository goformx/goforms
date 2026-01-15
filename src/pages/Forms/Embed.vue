<script setup lang="ts">
import { ref, computed } from "vue";
import { Link } from "@inertiajs/vue3";
import DashboardLayout from "@/components/layout/DashboardLayout.vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import {
  Pencil,
  Copy,
  Check,
  AlertCircle,
  CheckCircle2,
  Code,
  Terminal,
  Globe,
  Key,
  Link as LinkIcon,
} from "lucide-vue-next";
import { toast } from "vue-sonner";

interface Form {
  id: string;
  title: string;
  description: string;
  status: "draft" | "published" | "archived";
  corsOrigins: string[];
}

interface Props {
  form: Form;
  apiBaseUrl: string;
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = defineProps<Props>();

const copiedField = ref<string | null>(null);

// Computed values
const isPublished = computed(() => props.form.status === "published");
const hasCorsConfigured = computed(() => 
  props.form.corsOrigins && props.form.corsOrigins.length > 0
);
const apiEndpoints = computed(() => ({
  schema: `${props.apiBaseUrl}/api/v1/forms/${props.form.id}/schema`,
  validation: `${props.apiBaseUrl}/api/v1/forms/${props.form.id}/validation`,
  submit: `${props.apiBaseUrl}/api/v1/forms/${props.form.id}/submit`,
}));

// Code snippets
const svelteCode = computed(() => `<script lang="ts">
  import { useGoFormXForm } from '$lib/components/composables/useGoFormXForm.svelte';

  interface MyFormFields {
    email: string;
    name: string;
    message: string;
    [key: string]: string;
  }

  const form = useGoFormXForm<MyFormFields>({
    formId: '${props.form.id}',
    initialValues: {
      email: '',
      name: '',
      message: ''
    },
    onSuccess: (response) => {
      console.log('Form submitted!', response);
    }
  });
<\/script>

<form onsubmit={form.handleSubmit}>
  {#if form.isError}
    <div class="error">{form.errorMessage}</div>
  {/if}
  
  <input
    type="text"
    name="name"
    bind:value={form.fields.name}
    placeholder="Your name"
    required
  />
  
  <input
    type="email"
    name="email"
    bind:value={form.fields.email}
    placeholder="your@email.com"
    required
  />
  
  <textarea
    name="message"
    bind:value={form.fields.message}
    placeholder="Your message"
    required
  ></textarea>
  
  <button type="submit" disabled={form.isSubmitting}>
    {form.isSubmitting ? 'Sending...' : 'Submit'}
  </button>
  
  {#if form.isSuccess}
    <p class="success">Thank you for your submission!</p>
  {/if}
</form>`);

const vanillaJsCode = computed(() => `// Fetch form schema (optional - for dynamic form generation)
const schemaResponse = await fetch('${apiEndpoints.value.schema}', {
  headers: {
    'X-API-Key': 'YOUR_API_KEY',
    'Accept': 'application/json'
  }
});
const schema = await schemaResponse.json();

// Submit form data
const formData = {
  name: document.getElementById('name').value,
  email: document.getElementById('email').value,
  message: document.getElementById('message').value
};

const response = await fetch('${apiEndpoints.value.submit}', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'YOUR_API_KEY'
  },
  body: JSON.stringify(formData)
});

const result = await response.json();
if (result.success) {
  console.log('Submission ID:', result.submission_id);
} else {
  console.error('Error:', result.message);
}`);

const curlCode = computed(() => `# Get form schema
curl -X GET '${apiEndpoints.value.schema}' \\
  -H 'X-API-Key: YOUR_API_KEY' \\
  -H 'Accept: application/json'

# Submit form data
curl -X POST '${apiEndpoints.value.submit}' \\
  -H 'Content-Type: application/json' \\
  -H 'X-API-Key: YOUR_API_KEY' \\
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "message": "Hello from the API!"
  }'`);

// Copy to clipboard
async function copyToClipboard(text: string, field: string) {
  try {
    await navigator.clipboard.writeText(text);
    copiedField.value = field;
    toast.success("Copied to clipboard!");
    setTimeout(() => {
      copiedField.value = null;
    }, 2000);
  } catch {
    toast.error("Failed to copy to clipboard");
  }
}

// Show flash messages
if (props.flash?.success) {
  toast.success(props.flash.success);
}
if (props.flash?.error) {
  toast.error(props.flash.error);
}
</script>

<template>
  <DashboardLayout :title="`Embed: ${props.form.title}`" subtitle="Integrate this form into your website or application">
    <template #actions>
      <Button variant="outline" as-child>
        <Link :href="`/forms/${props.form.id}/edit`">
          <Pencil class="mr-2 h-4 w-4" />
          Edit Form
        </Link>
      </Button>
    </template>

    <div class="space-y-6">
      <!-- Status Alerts -->
      <Alert v-if="!isPublished" variant="warning">
        <AlertCircle class="h-4 w-4" />
        <AlertTitle>Form Not Published</AlertTitle>
        <AlertDescription>
          This form is currently in <Badge variant="secondary">{{ props.form.status }}</Badge> status. 
          Publish it to accept submissions from external websites.
          <Link :href="`/forms/${props.form.id}/edit`" class="underline ml-1">
            Edit form settings
          </Link>
        </AlertDescription>
      </Alert>

      <Alert v-if="isPublished && !hasCorsConfigured" variant="warning">
        <Globe class="h-4 w-4" />
        <AlertTitle>CORS Not Configured</AlertTitle>
        <AlertDescription>
          No allowed origins are configured. Add your website domain(s) to enable form submissions.
          <Link :href="`/forms/${props.form.id}/edit`" class="underline ml-1">
            Configure CORS
          </Link>
        </AlertDescription>
      </Alert>

      <Alert v-if="isPublished && hasCorsConfigured" variant="success">
        <CheckCircle2 class="h-4 w-4" />
        <AlertTitle>Ready for Integration</AlertTitle>
        <AlertDescription>
          This form is published and configured to accept submissions from: 
          <span class="font-mono text-sm">{{ props.form.corsOrigins.join(', ') }}</span>
        </AlertDescription>
      </Alert>

      <div class="grid gap-6 lg:grid-cols-3">
        <!-- Form ID Card -->
        <Card class="bg-card/50 backdrop-blur-sm border-border/50">
          <CardHeader class="pb-3">
            <CardTitle class="text-base flex items-center gap-2">
              <Key class="h-4 w-4" />
              Form ID
            </CardTitle>
            <CardDescription>Use this ID in your integration code</CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 bg-muted rounded-md text-sm font-mono truncate">
                {{ props.form.id }}
              </code>
              <Button
                variant="ghost"
                size="icon"
                @click="copyToClipboard(props.form.id, 'formId')"
              >
                <Check v-if="copiedField === 'formId'" class="h-4 w-4 text-green-500" />
                <Copy v-else class="h-4 w-4" />
              </Button>
            </div>
          </CardContent>
        </Card>

        <!-- API Endpoints Card -->
        <Card class="lg:col-span-2 bg-card/50 backdrop-blur-sm border-border/50">
          <CardHeader class="pb-3">
            <CardTitle class="text-base flex items-center gap-2">
              <LinkIcon class="h-4 w-4" />
              API Endpoints
            </CardTitle>
            <CardDescription>REST endpoints for this form</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center gap-2">
              <Badge variant="outline" class="shrink-0 w-14 justify-center">GET</Badge>
              <code class="flex-1 px-2 py-1 bg-muted rounded text-xs font-mono truncate">
                {{ apiEndpoints.schema }}
              </code>
              <Button
                variant="ghost"
                size="icon"
                class="shrink-0 h-8 w-8"
                @click="copyToClipboard(apiEndpoints.schema, 'schema')"
              >
                <Check v-if="copiedField === 'schema'" class="h-3 w-3 text-green-500" />
                <Copy v-else class="h-3 w-3" />
              </Button>
            </div>
            <div class="flex items-center gap-2">
              <Badge variant="outline" class="shrink-0 w-14 justify-center">GET</Badge>
              <code class="flex-1 px-2 py-1 bg-muted rounded text-xs font-mono truncate">
                {{ apiEndpoints.validation }}
              </code>
              <Button
                variant="ghost"
                size="icon"
                class="shrink-0 h-8 w-8"
                @click="copyToClipboard(apiEndpoints.validation, 'validation')"
              >
                <Check v-if="copiedField === 'validation'" class="h-3 w-3 text-green-500" />
                <Copy v-else class="h-3 w-3" />
              </Button>
            </div>
            <div class="flex items-center gap-2">
              <Badge variant="default" class="shrink-0 w-14 justify-center">POST</Badge>
              <code class="flex-1 px-2 py-1 bg-muted rounded text-xs font-mono truncate">
                {{ apiEndpoints.submit }}
              </code>
              <Button
                variant="ghost"
                size="icon"
                class="shrink-0 h-8 w-8"
                @click="copyToClipboard(apiEndpoints.submit, 'submit')"
              >
                <Check v-if="copiedField === 'submit'" class="h-3 w-3 text-green-500" />
                <Copy v-else class="h-3 w-3" />
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>

      <Separator />

      <!-- Code Snippets -->
      <Card class="bg-card/50 backdrop-blur-sm border-border/50">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Code class="h-5 w-5" />
            Integration Code
          </CardTitle>
          <CardDescription>
            Copy these code snippets to integrate the form into your application
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs default-value="svelte" class="w-full">
            <TabsList class="grid w-full grid-cols-3 lg:w-auto lg:inline-grid">
              <TabsTrigger value="svelte">Svelte 5</TabsTrigger>
              <TabsTrigger value="javascript">JavaScript</TabsTrigger>
              <TabsTrigger value="curl">cURL</TabsTrigger>
            </TabsList>

            <TabsContent value="svelte" class="mt-4">
              <div class="relative">
                <Button
                  variant="ghost"
                  size="sm"
                  class="absolute top-2 right-2 z-10"
                  @click="copyToClipboard(svelteCode, 'svelte')"
                >
                  <Check v-if="copiedField === 'svelte'" class="h-4 w-4 mr-1 text-green-500" />
                  <Copy v-else class="h-4 w-4 mr-1" />
                  Copy
                </Button>
                <pre class="p-4 bg-muted rounded-lg overflow-x-auto text-sm"><code>{{ svelteCode }}</code></pre>
              </div>
              <p class="mt-3 text-sm text-muted-foreground">
                Install the composable in your SvelteKit project and customize the form fields to match your needs.
              </p>
            </TabsContent>

            <TabsContent value="javascript" class="mt-4">
              <div class="relative">
                <Button
                  variant="ghost"
                  size="sm"
                  class="absolute top-2 right-2 z-10"
                  @click="copyToClipboard(vanillaJsCode, 'javascript')"
                >
                  <Check v-if="copiedField === 'javascript'" class="h-4 w-4 mr-1 text-green-500" />
                  <Copy v-else class="h-4 w-4 mr-1" />
                  Copy
                </Button>
                <pre class="p-4 bg-muted rounded-lg overflow-x-auto text-sm"><code>{{ vanillaJsCode }}</code></pre>
              </div>
              <p class="mt-3 text-sm text-muted-foreground">
                Use the Fetch API directly for framework-agnostic integration. Works with React, Vue, Angular, or vanilla JS.
              </p>
            </TabsContent>

            <TabsContent value="curl" class="mt-4">
              <div class="relative">
                <Button
                  variant="ghost"
                  size="sm"
                  class="absolute top-2 right-2 z-10"
                  @click="copyToClipboard(curlCode, 'curl')"
                >
                  <Check v-if="copiedField === 'curl'" class="h-4 w-4 mr-1 text-green-500" />
                  <Copy v-else class="h-4 w-4 mr-1" />
                  Copy
                </Button>
                <pre class="p-4 bg-muted rounded-lg overflow-x-auto text-sm"><code>{{ curlCode }}</code></pre>
              </div>
              <p class="mt-3 text-sm text-muted-foreground">
                Test the API directly from your terminal or use these commands as a reference for server-side integration.
              </p>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      <!-- API Key Notice -->
      <Alert>
        <Terminal class="h-4 w-4" />
        <AlertTitle>API Key Required</AlertTitle>
        <AlertDescription>
          Replace <code class="px-1 py-0.5 bg-muted rounded text-xs">YOUR_API_KEY</code> with your actual API key. 
          Generate one in your 
          <Link href="/dashboard/settings" class="underline">account settings</Link>.
        </AlertDescription>
      </Alert>
    </div>
  </DashboardLayout>
</template>
