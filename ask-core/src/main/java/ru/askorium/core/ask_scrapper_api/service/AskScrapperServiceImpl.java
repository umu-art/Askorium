package ru.askorium.core.ask_scrapper_api.service;

import lombok.RequiredArgsConstructor;
import org.redisson.api.RedissonClient;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Service;
import ru.askorium.api.model.ScrappedPage;
import ru.askorium.api.model.ScrapperResult;
import ru.askorium.core.ask_scrapper_api.AskScrapperService;
import ru.askorium.core.ask_scrapper_api.config.AmqpProperties;

import java.util.List;

@Service
@RequiredArgsConstructor
public class AskScrapperServiceImpl implements AskScrapperService {

    private final AmqpProperties amqpProperties;
    private final RabbitTemplate rabbitTemplate;
    private final RedissonClient redissonClient;

    @Override
    public List<ScrappedPage> scrapSource(String sourceUrl) {
        return List.of(); // TODO: complete
    }

    @RabbitListener(queues = "${askorium.amqp.scrapped-pages-queue}")
    public void handleScrappedPage(ScrapperResult scrapperResult) {
    }
}
