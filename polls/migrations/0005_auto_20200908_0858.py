# Generated by Django 3.1.1 on 2020-09-08 08:58

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('polls', '0004_auto_20200908_0631'),
    ]

    operations = [
        migrations.AddField(
            model_name='group',
            name='group_rank',
            field=models.IntegerField(default=0),
        ),
        migrations.AddField(
            model_name='player',
            name='player_rank',
            field=models.IntegerField(default=0),
        ),
    ]
