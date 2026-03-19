import { PUBLIC_API_BASE_URL } from "$env/static/public";
import { json } from "@sveltejs/kit";

function resolveBackendHealthURL(): string | null {
	if (!PUBLIC_API_BASE_URL) {
		return null;
	}

	return new URL("/api/v1/health", PUBLIC_API_BASE_URL).toString();
}

export async function GET({ fetch }) {
	const backendHealthURL = resolveBackendHealthURL();
	if (!backendHealthURL) {
		return json(
			{
				status: "error",
				message: "PUBLIC_API_BASE_URL is not configured",
			},
			{ status: 503 },
		);
	}

	try {
		const response = await fetch(backendHealthURL, {
			method: "GET",
		});

		if (!response.ok) {
			return json(
				{
					status: "error",
					message: "backend healthcheck failed",
					api_base_url: PUBLIC_API_BASE_URL,
					backend_health_url: backendHealthURL,
					backend_status: response.status,
				},
				{ status: 503 },
			);
		}
	} catch (error) {
		return json(
			{
				status: "error",
				message: "backend healthcheck unreachable",
				api_base_url: PUBLIC_API_BASE_URL,
				backend_health_url: backendHealthURL,
				error: error instanceof Error ? error.message : "unknown error",
			},
			{ status: 503 },
		);
	}

	return json({
		status: "ok",
		api_base_url: PUBLIC_API_BASE_URL,
		backend_health_url: backendHealthURL,
	});
}
