from django.db import models

# Create your models here.
# InnerMatch
class Group(models.Model):
    match_id = models.CharField(max_length=20, default="20200101")
    group_id = models.IntegerField(default=0)
    group_rank = models.IntegerField(default=0)
    group_name = models.CharField(max_length=200)
    group_join_rate = models.IntegerField(default=0)
    group_avg_score = models.IntegerField(default=0)
    def __str__(self):
        return self.group_name

class Player(models.Model):
    match_id = models.CharField(max_length=20, default="20200101")
    player_rank = models.IntegerField(default=0)
    player_uid = models.CharField(max_length=200)
    player_name = models.CharField(max_length=200)
    player_total_score = models.IntegerField(default=0)
    player_join_num = models.IntegerField(default=0)
    player_join_rate = models.IntegerField(default=0)
    player_group_id = models.IntegerField(default=0)
    player_total_game_time = models.IntegerField(default=0)
    def __str__(self):
        return self.player_name

class MatchID(models.Model):
    match_id = models.CharField(max_length=20, default="20200101")
    def __str__(self):
        return self.match_id

#example
class Question(models.Model):
    question_text = models.CharField(max_length=200)
    pub_date = models.DateTimeField('date published')
    def __str__(self):
        return self.question_text

    def was_published_recently(self):
        return self.pub_date >= timezone.now() - datetime.timedelta(days=1)


class Choice(models.Model):
    question = models.ForeignKey(Question, on_delete=models.CASCADE)
    choice_text = models.CharField(max_length=200)
    votes = models.IntegerField(default=0)
    def __str__(self):
        return self.choice_text
