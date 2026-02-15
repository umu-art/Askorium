package ru.askorium.core.user.config.interceptors;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.lang.NonNull;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.context.request.RequestAttributes;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.filter.OncePerRequestFilter;
import ru.askorium.core.user.domain.UserEntity;
import ru.askorium.core.user.jpa.UserJpa;

import java.io.IOException;
import java.time.OffsetDateTime;
import java.util.Arrays;
import java.util.UUID;
import java.util.stream.Collectors;

import static ru.askorium.core.common.UserUtils.USER_ID_KEY;

@Component
@RequiredArgsConstructor
public class AuthenticationFilter extends OncePerRequestFilter {

    private static final String COOKIE_NAME = "ask_uid";
    private static final int MAX_AGE_SECONDS = 315_360_000; // ~10 years

    private final UserJpa userJpa;

    @Override
    @Transactional
    protected void doFilterInternal(
            @NonNull HttpServletRequest request,
            @NonNull HttpServletResponse response,
            @NonNull FilterChain filterChain
    ) throws ServletException, IOException {

        var needSetCookie = false;
        UUID userId = null;

        var cookieUid = readCookieUid(request);
        if (cookieUid != null) {
            var userOpt = userJpa.findById(cookieUid);
            if (userOpt.isPresent()) {
                var user = userOpt.get();
                user.setLastSeenAt(OffsetDateTime.now());
                user.setLastSeenIp(resolveClientIp(request));
                userJpa.save(user);
                userId = user.getId();
            }
        }

        if (userId == null) {
            var user = new UserEntity();
            user.setLastSeenAt(OffsetDateTime.now());
            user.setLastSeenIp(resolveClientIp(request));
            user.setFirstVisitUserAgent(request.getHeader("User-Agent"));
            user.setFirstVisitHeaders(collectHeaders(request));
            user = userJpa.save(user);
            userId = user.getId();
            needSetCookie = true;
        }

        setUserIdToContext(userId);

        filterChain.doFilter(request, response);

        if (needSetCookie) {
            response.addHeader("Set-Cookie",
                    COOKIE_NAME + "=" + userId
                            + "; HttpOnly; Secure; Path=/; Max-Age=" + MAX_AGE_SECONDS);
        }
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

    private String collectHeaders(HttpServletRequest request) {
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
