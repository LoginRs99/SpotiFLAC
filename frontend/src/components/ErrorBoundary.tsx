import { Component, type ErrorInfo, type ReactNode } from "react";
import { Button } from "@/components/ui/button";

interface ErrorBoundaryProps {
    children: ReactNode;
}

interface ErrorBoundaryState {
    error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
    state: ErrorBoundaryState = {
        error: null,
    };

    static getDerivedStateFromError(error: Error): ErrorBoundaryState {
        return { error };
    }

    componentDidCatch(error: Error, info: ErrorInfo) {
        console.error("SpotiFLAC UI crashed:", error, info.componentStack);
    }

    render() {
        if (!this.state.error) {
            return this.props.children;
        }

        return (<div className="min-h-screen bg-background text-foreground flex items-center justify-center p-6">
            <div className="w-full max-w-xl rounded-lg border bg-card p-6 shadow-sm space-y-4">
                <div>
                    <h1 className="text-lg font-semibold">Something went wrong</h1>
                    <p className="mt-2 text-sm text-muted-foreground">
                        The Docker web UI hit a browser-side error. Refreshing may recover the session; the console will contain the detailed error.
                    </p>
                </div>
                <pre className="max-h-48 overflow-auto rounded-md bg-muted p-3 text-xs whitespace-pre-wrap">
                    {this.state.error.message}
                </pre>
                <Button onClick={() => window.location.reload()}>
                    Refresh
                </Button>
            </div>
        </div>);
    }
}
