export function shouldShowDesktopNavigation(): boolean {
    return !__DOCKER_WEB__;
}
