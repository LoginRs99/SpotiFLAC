import { lazy, type ComponentType } from "react";

type PageComponent = ComponentType<Record<string, never>>;

const EmptyPage: PageComponent = () => null;

const dockerHiddenPages = new Set([
    "debug",
    "projects",
    "support",
    "audio-analysis",
    "audio-converter",
    "audio-resampler",
    "file-manager",
]);

function optionalPage(loader: () => Promise<{ default: PageComponent }>): PageComponent {
    return __DOCKER_WEB__ ? EmptyPage : lazy(loader);
}

export function isDockerWebHiddenPage(page: string): boolean {
    return __DOCKER_WEB__ && dockerHiddenPages.has(page);
}

export const optionalPages = {
    DebugLoggerPage: optionalPage(() => import("@/components/DebugLoggerPage").then((module) => ({ default: module.DebugLoggerPage }))),
    OtherProjects: optionalPage(() => import("@/components/OtherProjects").then((module) => ({ default: module.OtherProjects }))),
    SupportPage: optionalPage(() => import("@/components/SupportPage").then((module) => ({ default: module.SupportPage }))),
    AudioAnalysisPage: optionalPage(() => import("@/components/AudioAnalysisPage").then((module) => ({ default: module.AudioAnalysisPage }))),
    AudioConverterPage: optionalPage(() => import("@/components/AudioConverterPage").then((module) => ({ default: module.AudioConverterPage }))),
    AudioResamplerPage: optionalPage(() => import("@/components/AudioResamplerPage").then((module) => ({ default: module.AudioResamplerPage }))),
    FileManagerPage: optionalPage(() => import("@/components/FileManagerPage").then((module) => ({ default: module.FileManagerPage }))),
};
