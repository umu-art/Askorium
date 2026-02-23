package ru.askorium.core.user.config.interceptors;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import org.redisson.api.RateType;
import org.redisson.api.RedissonClient;
import org.springframework.lang.NonNull;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;
import ru.askorium.core.user.config.UserProperties;

import java.io.IOException;
import java.time.Duration;

import static ru.askorium.core.common.UserUtils.getUserId;

@Component
@RequiredArgsConstructor
public class RateLimitFilter extends OncePerRequestFilter {

    private final RedissonClient redissonClient;
    private final UserProperties userProperties;

    @Override
    protected void doFilterInternal(
            @NonNull HttpServletRequest request,
            @NonNull HttpServletResponse response,
            @NonNull FilterChain filterChain
    ) throws ServletException, IOException {

        var userId = getUserId();

        var limiter = redissonClient.getRateLimiter("rate-limit:user:" + userId);

        limiter.trySetRate(
                RateType.OVERALL,
                userProperties.getRateLimit().getRequestsPerMinute(),
                Duration.ofMinutes(1)
        );

        if (!limiter.tryAcquire()) {
            response.setStatus(429);
            response.addHeader("Retry-After", "60");
            return;
        }

        filterChain.doFilter(request, response);
    }
}
