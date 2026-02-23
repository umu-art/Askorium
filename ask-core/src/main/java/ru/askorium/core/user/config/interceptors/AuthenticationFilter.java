package ru.askorium.core.user.config.interceptors;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.lang.NonNull;
import org.springframework.stereotype.Component;
import org.springframework.web.context.request.RequestAttributes;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.filter.OncePerRequestFilter;
import ru.askorium.core.user.service.UserService;

import java.io.IOException;
import java.util.Arrays;
import java.util.UUID;
import java.util.stream.Collectors;

import static ru.askorium.core.common.UserUtils.USER_ID_KEY;

@Component
@RequiredArgsConstructor
public class AuthenticationFilter extends OncePerRequestFilter {

    private static final String COOKIE_NAME = "ask_uid";
    private static final int MAX_AGE_SECONDS = 315_360_000; // ~10 years

    private final UserService userService;

    @Override
    protected void doFilterInternal(
            @NonNull HttpServletRequest request,
            @NonNull HttpServletResponse response,
            @NonNull FilterChain filterChain
    ) throws ServletException, IOException {

        if (!request.getRequestURI().startsWith("/ask/")) {
            filterChain.doFilter(request, response);
            return;
        }

        var resolution = userService.resolveOrCreate(
                readCookieUid(request),
                resolveClientIp(request),
                request.getHeader("User-Agent"),
                collectParams(request)
        );

        if (resolution.isNew()) {
            response.addHeader("Set-Cookie",
                    COOKIE_NAME + "=" + resolution.userId()
                            + "; HttpOnly; Path=/; Max-Age=" + MAX_AGE_SECONDS);
        }

        setUserIdToContext(resolution.userId());

        filterChain.doFilter(request, response);
    }

    private UUID readCookieUid(HttpServletRequest request) {
        var cookies = request.getCookies();
        if (cookies == null) {
            return null;
        }

        return Arrays.stream(cookies)
                .filter(c -> COOKIE_NAME.equals(c.getName()))
                .map(Cookie::getValue)
                .findFirst()
                .map(val -> {
                    try {
                        return UUID.fromString(val);
                    } catch (IllegalArgumentException e) {
                        return null;
                    }
                })
                .orElse(null);
    }

    private String resolveClientIp(HttpServletRequest request) {
        String forwarded = request.getHeader("X-Forwarded-For");
        if (forwarded != null && !forwarded.isBlank()) {
            return forwarded.split(",")[0].trim();
        }
        return request.getRemoteAddr();
    }

    private String collectParams(HttpServletRequest request) {
        return request.getParameterMap()
                .entrySet()
                .stream()
                .map(e -> e.getKey() + "=" + String.join(",", e.getValue()))
                .collect(Collectors.joining(";"));
    }

    private void setUserIdToContext(UUID userId) {
        RequestContextHolder.currentRequestAttributes()
                .setAttribute(USER_ID_KEY, userId, RequestAttributes.SCOPE_REQUEST);
    }
}
